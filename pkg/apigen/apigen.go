// Package apigen provides functionality to automatically generate JSON-friendly API descriptions
// from Go source code. It analyzes Go packages and files to extract method signatures,
// parameter types, struct fields, and struct tag annotations to create comprehensive
// API documentation that can be consumed by client libraries or used for runtime
// API introspection.
//
// Key features:
//   - Parse Go packages or individual files
//   - Extract method descriptions from Go doc comments
//   - Analyze parameter types including complex structs
//   - Parse struct tags and include them as annotations
//   - Filter methods using various strategies (prefix, suffix, contains, explicit list)
//   - Generate Go files with embedded API definitions for runtime use
//   - CLI tool for easy integration with go:generate
//
// Example usage:
//
//	parser := apigen.NewParser()
//	transformer := apigen.NewTransformer()
//	generator := apigen.NewJSONGenerator()
//
//	methods, err := parser.ParsePackage("./pkg/handlers")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	filtered := apigen.FilterByPrefix(methods, "Handle")
//	enriched, err := transformer.Transform(filtered)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	desc := apigen.NewDescription("MyAPI", enriched)
//	content, err := generator.Generate(desc)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// For go:generate integration, see the cmd/apigen CLI tool.
package apigen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// DefaultParser is the default implementation of Parser
type DefaultParser struct {
	fset *token.FileSet
}

// NewParser creates a new parser instance
func NewParser() Parser {
	return &DefaultParser{
		fset: token.NewFileSet(),
	}
}

// ParsePackage parses all Go files in a package directory
func (p *DefaultParser) ParsePackage(packagePath string) ([]RawMethod, error) {
	pkgs, err := parser.ParseDir(p.fset, packagePath, func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package: %w", err)
	}

	var allMethods []RawMethod
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			methods, err := p.parseFile(file)
			if err != nil {
				return nil, fmt.Errorf("failed to parse file %s: %w", p.fset.File(file.Pos()).Name(), err)
			}
			allMethods = append(allMethods, methods...)
		}
	}

	return allMethods, nil
}

// ParseSingleFile parses a single Go file
func (p *DefaultParser) ParseSingleFile(filePath string) ([]RawMethod, error) {
	file, err := parser.ParseFile(p.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	return p.parseFile(file)
}

// parseFile parses methods from an AST file
func (p *DefaultParser) parseFile(file *ast.File) ([]RawMethod, error) {
	var methods []RawMethod

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		method, err := p.parseMethod(funcDecl)
		if err != nil {
			return nil, fmt.Errorf("failed to parse method %s: %w", funcDecl.Name.Name, err)
		}
		methods = append(methods, method)
	}

	return methods, nil
}

// parseMethod converts an AST function declaration to a RawMethod
func (p *DefaultParser) parseMethod(funcDecl *ast.FuncDecl) (RawMethod, error) {
	method := RawMethod{
		Name:     funcDecl.Name.Name,
		Position: p.fset.Position(funcDecl.Pos()),
		Doc:      p.extractDocComments(funcDecl),
	}

	// Parse parameters
	if funcDecl.Type.Params != nil {
		for _, field := range funcDecl.Type.Params.List {
			params, err := p.parseFieldList(field)
			if err != nil {
				return method, fmt.Errorf("failed to parse parameters: %w", err)
			}
			method.Params = append(method.Params, params...)
		}
	}

	return method, nil
}

// parseFieldList converts AST field list to RawParams
func (p *DefaultParser) parseFieldList(field *ast.Field) ([]RawParam, error) {
	var params []RawParam

	// Handle multiple parameter names with same type
	if len(field.Names) == 0 {
		// Anonymous parameter (shouldn't happen in valid Go, but handle gracefully)
		return params, nil
	}

	for _, name := range field.Names {
		params = append(params, RawParam{
			Name: name.Name,
			Type: field.Type,
		})
	}

	return params, nil
}

// extractDocComments extracts documentation comments from a function
func (p *DefaultParser) extractDocComments(funcDecl *ast.FuncDecl) []string {
	if funcDecl.Doc == nil || len(funcDecl.Doc.List) == 0 {
		return nil
	}

	var comments []string
	for _, comment := range funcDecl.Doc.List {
		text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
		if text != "" {
			comments = append(comments, text)
		}
	}

	return comments
}

// DefaultTransformer is the default implementation of Transformer
type DefaultTransformer struct {
	registry *TypeRegistry
}

// NewTransformer creates a new transformer with the given type registry
func NewTransformer(registry *TypeRegistry) Transformer {
	return &DefaultTransformer{
		registry: registry,
	}
}

// Transform converts RawMethods to EnrichedMethods with resolved types
func (t *DefaultTransformer) Transform(methods []RawMethod) ([]EnrichedMethod, error) {
	var enriched []EnrichedMethod

	for _, method := range methods {
		enrichedMethod, err := t.transformMethod(method)
		if err != nil {
			return nil, fmt.Errorf("failed to transform method %s: %w", method.Name, err)
		}
		enriched = append(enriched, enrichedMethod)
	}

	return enriched, nil
}

// transformMethod converts a RawMethod to an EnrichedMethod
func (t *DefaultTransformer) transformMethod(method RawMethod) (EnrichedMethod, error) {
	enriched := EnrichedMethod{
		Name:        method.Name,
		Description: strings.Join(method.Doc, " "),
	}

	for _, param := range method.Params {
		enrichedParam, err := t.transformParam(param)
		if err != nil {
			return enriched, fmt.Errorf("failed to transform parameter %s: %w", param.Name, err)
		}
		enriched.Parameters = append(enriched.Parameters, enrichedParam)
	}

	return enriched, nil
}

// transformParam converts a RawParam to an EnrichedParam
func (t *DefaultTransformer) transformParam(param RawParam) (EnrichedParam, error) {
	parsedType, err := t.parseType(param.Type)
	if err != nil {
		return EnrichedParam{}, err
	}

	enriched := EnrichedParam{
		Name: param.Name,
		Type: *parsedType,
	}

	// If this is a struct type defined in our registry, resolve the field details
	if parsedType.Kind == TypeKindBasic || (parsedType.Kind == TypeKindSelector && parsedType.Package == "") {
		if fieldType, exists := t.registry.GetType(parsedType.Name); exists && fieldType.Kind == TypeKindStruct {
			enriched.Field = &ParsedField{
				Name: param.Name,
				Type: *fieldType,
			}
		}
	}

	return enriched, nil
}

// parseType converts an AST expression to a ParsedType
func (t *DefaultTransformer) parseType(expr ast.Expr) (*ParsedType, error) {
	switch e := expr.(type) {
	case *ast.Ident:
		return &ParsedType{
			Kind: TypeKindBasic,
			Name: e.Name,
		}, nil

	case *ast.SelectorExpr:
		pkg, ok := e.X.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("invalid selector expression")
		}
		return &ParsedType{
			Kind:    TypeKindSelector,
			Name:    e.Sel.Name,
			Package: pkg.Name,
		}, nil

	case *ast.StarExpr:
		elemType, err := t.parseType(e.X)
		if err != nil {
			return nil, err
		}
		elemType.IsPointer = true
		return elemType, nil

	case *ast.ArrayType:
		elemType, err := t.parseType(e.Elt)
		if err != nil {
			return nil, err
		}
		elemType.IsSlice = true
		return elemType, nil

	case *ast.MapType:
		keyType, err := t.parseType(e.Key)
		if err != nil {
			return nil, err
		}
		valueType, err := t.parseType(e.Value)
		if err != nil {
			return nil, err
		}
		return &ParsedType{
			Kind:      TypeKindMap,
			IsMap:     true,
			KeyType:   keyType,
			ValueType: valueType,
		}, nil

	default:
		return &ParsedType{
			Kind: TypeKindUnknown,
			Name: "unknown",
		}, nil
	}
}

// JSONGenerator generates JSON output
type JSONGenerator struct{}

// NewJSONGenerator creates a new JSON generator
func NewJSONGenerator() Generator {
	return &JSONGenerator{}
}

// Generate generates JSON representation of the API description
func (g *JSONGenerator) Generate(desc APIDescription) (GeneratedContent, error) {
	data, err := json.MarshalIndent(desc, "", "  ")
	if err != nil {
		return GeneratedContent{}, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return GeneratedContent{
		Content: string(data),
	}, nil
}

// GoConstGenerator generates Go constant declarations
type GoConstGenerator struct {
	packageName string
	constName   string
}

// NewGoConstGenerator creates a new Go constant generator
func NewGoConstGenerator(packageName, constName string) Generator {
	return &GoConstGenerator{
		packageName: packageName,
		constName:   constName,
	}
}

// Generate generates Go code with a constant containing JSON
func (g *GoConstGenerator) Generate(desc APIDescription) (GeneratedContent, error) {
	jsonData, err := json.MarshalIndent(desc, "", "  ")
	if err != nil {
		return GeneratedContent{}, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	tmpl, err := template.New("goConst").Parse(goConstTemplate)
	if err != nil {
		return GeneratedContent{}, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		PackageName string
		ConstName   string
		JSONContent string
	}{
		PackageName: g.packageName,
		ConstName:   g.constName,
		JSONContent: string(jsonData),
	})
	if err != nil {
		return GeneratedContent{}, fmt.Errorf("failed to execute template: %w", err)
	}

	return GeneratedContent{
		Content:     buf.String(),
		PackageName: g.packageName,
		ConstName:   g.constName,
	}, nil
}

// GoMapGenerator generates Go map declarations
type GoMapGenerator struct {
	packageName string
	varName     string
}

// NewGoMapGenerator creates a new Go map generator
func NewGoMapGenerator(packageName, varName string) Generator {
	return &GoMapGenerator{
		packageName: packageName,
		varName:     varName,
	}
}

// Generate generates Go code with a map of method names to JSON strings
func (g *GoMapGenerator) Generate(desc APIDescription) (GeneratedContent, error) {
	tmpl, err := template.New("goMap").Parse(goMapTemplate)
	if err != nil {
		return GeneratedContent{}, fmt.Errorf("failed to parse template: %w", err)
	}

	methods := make(map[string]string)
	for name, method := range desc.Methods {
		jsonData, err := json.MarshalIndent(method, "", "  ")
		if err != nil {
			return GeneratedContent{}, fmt.Errorf("failed to marshal method %s: %w", name, err)
		}
		methods[name] = string(jsonData)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		PackageName string
		VarName     string
		Methods     map[string]string
	}{
		PackageName: g.packageName,
		VarName:     g.varName,
		Methods:     methods,
	})
	if err != nil {
		return GeneratedContent{}, fmt.Errorf("failed to execute template: %w", err)
	}

	return GeneratedContent{
		Content:     buf.String(),
		PackageName: g.packageName,
		ConstName:   g.varName,
	}, nil
}

// DefaultWriter is the default implementation of Writer
type DefaultWriter struct{}

// NewWriter creates a new writer instance
func NewWriter() Writer {
	return &DefaultWriter{}
}

// WriteToFile writes generated content to a file
func (w *DefaultWriter) WriteToFile(content GeneratedContent, filePath string) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	err := os.WriteFile(filePath, []byte(content.Content), 0o644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// NewDescription creates an API description from enriched methods
func NewDescription(apiName string, methods []EnrichedMethod) APIDescription {
	desc := APIDescription{
		APIName: apiName,
		Methods: make(map[string]MethodDescription),
	}

	for _, method := range methods {
		methodDesc := MethodDescription{
			Description: method.Description,
			Parameters:  make(map[string]ParameterInfo),
		}

		for _, param := range method.Parameters {
			paramInfo := ParameterInfo{
				Type: param.Type.Name,
			}

			// Add fields if this parameter is a struct
			if param.Field != nil && len(param.Field.Type.Fields) > 0 {
				paramInfo.Fields = make(map[string]FieldInfo)
				for _, field := range param.Field.Type.Fields {
					fieldInfo := FieldInfo{
						Type:        field.Type.Name,
						Annotations: field.Tags,
					}
					paramInfo.Fields[field.Name] = fieldInfo
				}
			}

			methodDesc.Parameters[param.Name] = paramInfo
		}

		desc.Methods[method.Name] = methodDesc
	}

	return desc
}

// FilterByPrefix filters methods by prefix
func FilterByPrefix(methods []RawMethod, prefix string) []RawMethod {
	var filtered []RawMethod
	for _, method := range methods {
		if strings.HasPrefix(method.Name, prefix) {
			filtered = append(filtered, method)
		}
	}
	return filtered
}

// FilterBySuffix filters methods by suffix
func FilterBySuffix(methods []RawMethod, suffix string) []RawMethod {
	var filtered []RawMethod
	for _, method := range methods {
		if strings.HasSuffix(method.Name, suffix) {
			filtered = append(filtered, method)
		}
	}
	return filtered
}

// FilterByContains filters methods containing a substring
func FilterByContains(methods []RawMethod, substr string) []RawMethod {
	var filtered []RawMethod
	for _, method := range methods {
		if strings.Contains(method.Name, substr) {
			filtered = append(filtered, method)
		}
	}
	return filtered
}

// FilterByList filters methods by explicit list
func FilterByList(methods []RawMethod, names []string) []RawMethod {
	nameSet := make(map[string]bool)
	for _, name := range names {
		nameSet[name] = true
	}

	var filtered []RawMethod
	for _, method := range methods {
		if nameSet[method.Name] {
			filtered = append(filtered, method)
		}
	}
	return filtered
}

const goConstTemplate = `// Code generated by apigen; DO NOT EDIT.
// This file contains the API description for {{.PackageName}}

package {{.PackageName}}

// {{.ConstName}} contains the JSON API description
const {{.ConstName}} = ` + "`" + `{{.JSONContent}}` + "`" + `
`

const goMapTemplate = `// Code generated by apigen; DO NOT EDIT.
// This file contains the API description for {{.PackageName}}

package {{.PackageName}}

import "encoding/json"

// {{.VarName}} contains the API description as a map of method names to JSON strings
var {{.VarName}} = map[string]string{
{{- range $name, $json := .Methods}}
	"{{$name}}": ` + "`{{$json}}`" + `,
{{- end}}
}

// GetMethodDescription retrieves and parses a method description by name
func GetMethodDescription(name string) (MethodDescription, error) {
	jsonStr, exists := {{.VarName}}[name]
	if !exists {
		return MethodDescription{}, fmt.Errorf("method %s not found", name)
	}

	var desc MethodDescription
	err := json.Unmarshal([]byte(jsonStr), &desc)
	return desc, err
}
`