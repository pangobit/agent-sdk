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
	"strings"
	"text/template"
)

// DefaultParser is the default implementation of Parser
type DefaultParser struct {
	fset     *token.FileSet
	registry *TypeRegistry
}

// NewParser creates a new parser instance with an empty type registry
func NewParser() Parser {
	return &DefaultParser{
		fset:     token.NewFileSet(),
		registry: NewTypeRegistry(),
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

	// Resolve all type references after parsing all files
	p.registry.ResolveAllTypes()

	return allMethods, nil
}

// ParseSingleFile parses a single Go file
func (p *DefaultParser) ParseSingleFile(filePath string) ([]RawMethod, error) {
	file, err := parser.ParseFile(p.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	methods, err := p.parseFile(file)
	if err != nil {
		return nil, err
	}

	// Resolve all type references after parsing
	p.registry.ResolveAllTypes()

	return methods, nil
}

// parseFile parses methods and type declarations from an AST file
func (p *DefaultParser) parseFile(file *ast.File) ([]RawMethod, error) {
	var methods []RawMethod

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			method, err := p.parseMethod(d)
			if err != nil {
				return nil, fmt.Errorf("failed to parse method %s: %w", d.Name.Name, err)
			}
			methods = append(methods, method)

		case *ast.GenDecl:
			if d.Tok == token.TYPE {
				err := p.parseTypeDeclarations(d)
				if err != nil {
					return nil, fmt.Errorf("failed to parse type declarations: %w", err)
				}
			}
		}
	}

	return methods, nil
}

// parseTypeDeclarations parses type declarations from a GenDecl and adds them to the registry
func (p *DefaultParser) parseTypeDeclarations(genDecl *ast.GenDecl) error {
	for _, spec := range genDecl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		parsedType, err := p.parseType(typeSpec.Type)
		if err != nil {
			return fmt.Errorf("failed to parse type %s: %w", typeSpec.Name.Name, err)
		}

		parsedType.Name = typeSpec.Name.Name
		p.registry.AddType(typeSpec.Name.Name, parsedType)
	}

	return nil
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

// GetRegistry returns the type registry populated during parsing
func (p *DefaultParser) GetRegistry() *TypeRegistry {
	return p.registry
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
	resolver *TypeResolver
}

// NewTransformer creates a new transformer with the given type registry
func NewTransformer(registry *TypeRegistry) Transformer {
	return &DefaultTransformer{
		registry: registry,
		resolver: NewTypeResolver(registry),
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

	// Try to resolve the type using the resolver
	if parsedType.Kind == TypeKindBasic && parsedType.Name != "" {
		if resolvedType, err := t.resolver.ResolveType(parsedType.Name); err == nil {
			enriched.ResolvedType = resolvedType
		}
	} else {
		// For complex types, create a resolved version
		resolved, err := t.resolver.resolveParsedType(parsedType)
		if err == nil {
			enriched.ResolvedType = resolved
		}
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

// resolveType recursively resolves type references in the registry
func (t *DefaultTransformer) resolveType(pt *ParsedType) error {
	// Resolve named types
	if pt.Kind == TypeKindBasic && pt.Name != "" {
		if resolvedType, exists := t.registry.GetType(pt.Name); exists {
			// Copy the resolved type's fields
			pt.Fields = resolvedType.Fields
			pt.Kind = resolvedType.Kind
		}
	}

	// Recursively resolve nested types
	if pt.KeyType != nil {
		err := t.resolveType(pt.KeyType)
		if err != nil {
			return err
		}
	}
	if pt.ValueType != nil {
		err := t.resolveType(pt.ValueType)
		if err != nil {
			return err
		}
	}

	// Resolve fields
	for i := range pt.Fields {
		err := t.resolveType(&pt.Fields[i].Type)
		if err != nil {
			return err
		}
	}

	return nil
}

// parseType converts an AST expression to a ParsedType
func (t *DefaultTransformer) parseType(expr ast.Expr) (*ParsedType, error) {
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Name == "struct" {
			return &ParsedType{Kind: TypeKindStruct}, nil
		}
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
		// Create a new pointer type that wraps the element type
		return &ParsedType{
			Kind:      TypeKindPointer,
			IsPointer: true,
			KeyType:   elemType, // KeyType holds the pointed-to type for pointers
		}, nil

	case *ast.ArrayType:
		elemType, err := t.parseType(e.Elt)
		if err != nil {
			return nil, err
		}
		// Create a new slice type that wraps the element type
		return &ParsedType{
			Kind:    TypeKindSlice,
			IsSlice: true,
			KeyType: elemType, // KeyType holds the element type for slices
		}, nil

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

	case *ast.StructType:
		parsedType := &ParsedType{
			Kind: TypeKindStruct,
		}

		if e.Fields != nil {
			for _, field := range e.Fields.List {
				parsedFields, err := t.parseStructField(field)
				if err != nil {
					return nil, err
				}
				parsedType.Fields = append(parsedType.Fields, parsedFields...)
			}
		}

		return parsedType, nil

	default:
		return &ParsedType{
			Kind: TypeKindUnknown,
			Name: "unknown",
		}, nil
	}
}
func (p *DefaultParser) parseType(expr ast.Expr) (*ParsedType, error) {
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Name == "struct" {
			return &ParsedType{Kind: TypeKindStruct}, nil
		}
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
		elemType, err := p.parseType(e.X)
		if err != nil {
			return nil, err
		}
		// Create a new pointer type that wraps the element type
		return &ParsedType{
			Kind:      TypeKindPointer,
			IsPointer: true,
			KeyType:   elemType, // KeyType holds the pointed-to type for pointers
		}, nil

	case *ast.ArrayType:
		elemType, err := p.parseType(e.Elt)
		if err != nil {
			return nil, err
		}
		// Create a new slice type that wraps the element type
		return &ParsedType{
			Kind:    TypeKindSlice,
			IsSlice: true,
			KeyType: elemType, // KeyType holds the element type for slices
		}, nil

	case *ast.MapType:
		keyType, err := p.parseType(e.Key)
		if err != nil {
			return nil, err
		}
		valueType, err := p.parseType(e.Value)
		if err != nil {
			return nil, err
		}
		return &ParsedType{
			Kind:      TypeKindMap,
			IsMap:     true,
			KeyType:   keyType,
			ValueType: valueType,
		}, nil

	case *ast.StructType:
		parsedType := &ParsedType{
			Kind: TypeKindStruct,
		}

		if e.Fields != nil {
			for _, field := range e.Fields.List {
				parsedFields, err := p.parseStructField(field)
				if err != nil {
					return nil, err
				}
				parsedType.Fields = append(parsedType.Fields, parsedFields...)
			}
		}

		return parsedType, nil

	default:
		return &ParsedType{
			Kind: TypeKindUnknown,
			Name: "unknown",
		}, nil
	}
}

// parseStructField converts AST struct field to ParsedField
func (p *DefaultParser) parseStructField(field *ast.Field) ([]ParsedField, error) {
	fieldType, err := p.parseType(field.Type)
	if err != nil {
		return nil, err
	}

	var tags map[string]string
	if field.Tag != nil {
		tags, err = parseStructTags(field.Tag.Value)
		if err != nil {
			return nil, err
		}
	}

	var fields []ParsedField
	if len(field.Names) == 0 {
		// Embedded field
		fields = append(fields, ParsedField{
			Name: fieldType.Name,
			Type: *fieldType,
			Tags: tags,
		})
	} else {
		// Named fields
		for _, name := range field.Names {
			fields = append(fields, ParsedField{
				Name: name.Name,
				Type: *fieldType,
				Tags: tags,
			})
		}
	}

	return fields, nil
}

// parseStructField converts AST struct field to ParsedField
func (t *DefaultTransformer) parseStructField(field *ast.Field) ([]ParsedField, error) {
	fieldType, err := t.parseType(field.Type)
	if err != nil {
		return nil, err
	}

	var tags map[string]string
	if field.Tag != nil {
		tags, err = parseStructTags(field.Tag.Value)
		if err != nil {
			return nil, err
		}
	}

	var fields []ParsedField
	if len(field.Names) == 0 {
		// Embedded field
		fields = append(fields, ParsedField{
			Name: fieldType.Name,
			Type: *fieldType,
			Tags: tags,
		})
	} else {
		// Named fields
		for _, name := range field.Names {
			fields = append(fields, ParsedField{
				Name: name.Name,
				Type: *fieldType,
				Tags: tags,
			})
		}
	}

	return fields, nil
}

// parseStructTags parses Go struct tags into a map
func parseStructTags(tagStr string) (map[string]string, error) {
	annotations := make(map[string]string)

	// Remove surrounding quotes if present
	tagStr = strings.Trim(tagStr, "`")

	// Simple parser for struct tags
	parts := strings.Fields(tagStr)
	for _, part := range parts {
		// Split on first colon
		colonIndex := strings.Index(part, ":")
		if colonIndex == -1 {
			continue
		}

		key := part[:colonIndex]
		value := part[colonIndex+1:]

		// Remove surrounding quotes
		value = strings.Trim(value, `"`)

		annotations[key] = value
	}

	return annotations, nil
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

// NewDescription creates an API description from enriched methods
func NewDescription(apiName string, methods []EnrichedMethod) (APIDescription, error) {
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
			typeStr, err := typeToString(param.Type)
			if err != nil {
				return desc, fmt.Errorf("failed to stringify parameter %s type: %w", param.Name, err)
			}

			paramInfo := ParameterInfo{
				Type: typeStr,
			}

			// Use resolved type information if available
			if param.ResolvedType != nil {
				fields, err := buildFieldInfoFromResolved(param.ResolvedType)
				if err != nil {
					return desc, fmt.Errorf("failed to build resolved field info for parameter %s: %w", param.Name, err)
				}
				paramInfo.Fields = fields

				// Handle the parameter type itself being a slice or map
				if param.ResolvedType.IsSlice && param.ResolvedType.KeyType != nil {
					elementTypeStr, err := resolvedTypeToString(param.ResolvedType.KeyType)
					if err != nil {
						return desc, fmt.Errorf("failed to stringify element type for parameter %s: %w", param.Name, err)
					}

					elementTypeInfo := FieldInfo{
						Type: elementTypeStr,
					}
					elementTypeInfo.Fields, err = buildFieldInfoFromResolved(param.ResolvedType.KeyType)
					if err != nil {
						return desc, fmt.Errorf("failed to build element field info for parameter %s: %w", param.Name, err)
					}
					// For parameter-level slices/maps, we need to return this as the main info
					paramInfo.ElementType = &elementTypeInfo
				} else if param.ResolvedType.IsMap {
					if param.ResolvedType.KeyType != nil {
						keyTypeStr, err := resolvedTypeToString(param.ResolvedType.KeyType)
						if err != nil {
							return desc, fmt.Errorf("failed to stringify key type for parameter %s: %w", param.Name, err)
						}

						keyTypeInfo := FieldInfo{
							Type: keyTypeStr,
						}
						keyTypeInfo.Fields, err = buildFieldInfoFromResolved(param.ResolvedType.KeyType)
						if err != nil {
							return desc, fmt.Errorf("failed to build key field info for parameter %s: %w", param.Name, err)
						}
						paramInfo.KeyType = &keyTypeInfo
					}
					if param.ResolvedType.ValueType != nil {
						valueTypeStr, err := resolvedTypeToString(param.ResolvedType.ValueType)
						if err != nil {
							return desc, fmt.Errorf("failed to stringify value type for parameter %s: %w", param.Name, err)
						}

						valueTypeInfo := FieldInfo{
							Type: valueTypeStr,
						}
						valueTypeInfo.Fields, err = buildFieldInfoFromResolved(param.ResolvedType.ValueType)
						if err != nil {
							return desc, fmt.Errorf("failed to build value field info for parameter %s: %w", param.Name, err)
						}
						paramInfo.ValueType = &valueTypeInfo
					}
				}
			} else {
				// Fallback to old logic
				fields, err := buildFieldInfo(param.Type)
				if err != nil {
					return desc, fmt.Errorf("failed to build field info for parameter %s: %w", param.Name, err)
				}
				paramInfo.Fields = fields
			}

			methodDesc.Parameters[param.Name] = paramInfo
		}

		desc.Methods[method.Name] = methodDesc
	}

	return desc, nil
}

// buildFieldInfo recursively builds field information for a parsed type
func buildFieldInfo(pt ParsedType) (map[string]FieldInfo, error) {
	fields := make(map[string]FieldInfo)

	// If it's a struct, add its direct fields
	if pt.Kind == TypeKindStruct && len(pt.Fields) > 0 {
		for _, field := range pt.Fields {
			typeStr, err := typeToString(field.Type)
			if err != nil {
				return nil, fmt.Errorf("failed to stringify field %s type: %w", field.Name, err)
			}

			fieldInfo := FieldInfo{
				Type:        typeStr,
				Annotations: field.Tags,
			}
			// Recursively build nested fields
			nestedFields, err := buildFieldInfo(field.Type)
			if err != nil {
				return nil, fmt.Errorf("failed to build nested field info for %s: %w", field.Name, err)
			}
			fieldInfo.Fields = nestedFields

			// Handle nested slices and maps in struct fields
			if field.Type.IsSlice && field.Type.KeyType != nil {
				elementTypeStr, err := typeToString(*field.Type.KeyType)
				if err != nil {
					return nil, fmt.Errorf("failed to stringify slice element type for field %s: %w", field.Name, err)
				}

				elementTypeInfo := FieldInfo{
					Type: elementTypeStr,
				}
				elementTypeInfo.Fields, err = buildFieldInfo(*field.Type.KeyType)
				if err != nil {
					return nil, fmt.Errorf("failed to build element field info for field %s: %w", field.Name, err)
				}
				fieldInfo.ElementType = &elementTypeInfo
			}

			if field.Type.IsMap {
				if field.Type.KeyType != nil {
					keyTypeStr, err := typeToString(*field.Type.KeyType)
					if err != nil {
						return nil, fmt.Errorf("failed to stringify map key type for field %s: %w", field.Name, err)
					}

					keyTypeInfo := FieldInfo{
						Type: keyTypeStr,
					}
					keyTypeInfo.Fields, err = buildFieldInfo(*field.Type.KeyType)
					if err != nil {
						return nil, fmt.Errorf("failed to build key field info for field %s: %w", field.Name, err)
					}
					fieldInfo.KeyType = &keyTypeInfo
				}
				if field.Type.ValueType != nil {
					valueTypeStr, err := typeToString(*field.Type.ValueType)
					if err != nil {
						return nil, fmt.Errorf("failed to stringify map value type for field %s: %w", field.Name, err)
					}

					valueTypeInfo := FieldInfo{
						Type: valueTypeStr,
					}
					valueTypeInfo.Fields, err = buildFieldInfo(*field.Type.ValueType)
					if err != nil {
						return nil, fmt.Errorf("failed to build value field info for field %s: %w", field.Name, err)
					}
					fieldInfo.ValueType = &valueTypeInfo
				}
			}

			fields[field.Name] = fieldInfo
		}
	}

	return fields, nil
}

// buildFieldInfoFromResolved recursively builds field information from a ResolvedType
func buildFieldInfoFromResolved(rt *ResolvedType) (map[string]FieldInfo, error) {
	fields := make(map[string]FieldInfo)

	// If it's a struct, add its direct fields
	if rt.Kind == TypeKindStruct && len(rt.Fields) > 0 {
		for _, field := range rt.Fields {
			typeStr, err := resolvedTypeToString(field.Type)
			if err != nil {
				return nil, fmt.Errorf("failed to stringify resolved field %s type: %w", field.Name, err)
			}

			fieldInfo := FieldInfo{
				Type:        typeStr,
				Annotations: field.Tags,
			}
			// Recursively build nested fields
			nestedFields, err := buildFieldInfoFromResolved(field.Type)
			if err != nil {
				return nil, fmt.Errorf("failed to build nested resolved field info for %s: %w", field.Name, err)
			}
			fieldInfo.Fields = nestedFields

			// Handle nested slices and maps in struct fields
			if field.Type.IsSlice && field.Type.KeyType != nil {
				elementTypeStr, err := resolvedTypeToString(field.Type.KeyType)
				if err != nil {
					return nil, fmt.Errorf("failed to stringify resolved slice element type for field %s: %w", field.Name, err)
				}

				elementTypeInfo := FieldInfo{
					Type: elementTypeStr,
				}
				elementTypeInfo.Fields, err = buildFieldInfoFromResolved(field.Type.KeyType)
				if err != nil {
					return nil, fmt.Errorf("failed to build resolved element field info for field %s: %w", field.Name, err)
				}
				fieldInfo.ElementType = &elementTypeInfo
			}

			if field.Type.IsMap {
				if field.Type.KeyType != nil {
					keyTypeStr, err := resolvedTypeToString(field.Type.KeyType)
					if err != nil {
						return nil, fmt.Errorf("failed to stringify resolved map key type for field %s: %w", field.Name, err)
					}

					keyTypeInfo := FieldInfo{
						Type: keyTypeStr,
					}
					keyTypeInfo.Fields, err = buildFieldInfoFromResolved(field.Type.KeyType)
					if err != nil {
						return nil, fmt.Errorf("failed to build resolved key field info for field %s: %w", field.Name, err)
					}
					fieldInfo.KeyType = &keyTypeInfo
				}
				if field.Type.ValueType != nil {
					valueTypeStr, err := resolvedTypeToString(field.Type.ValueType)
					if err != nil {
						return nil, fmt.Errorf("failed to stringify resolved map value type for field %s: %w", field.Name, err)
					}

					valueTypeInfo := FieldInfo{
						Type: valueTypeStr,
					}
					valueTypeInfo.Fields, err = buildFieldInfoFromResolved(field.Type.ValueType)
					if err != nil {
						return nil, fmt.Errorf("failed to build resolved value field info for field %s: %w", field.Name, err)
					}
					fieldInfo.ValueType = &valueTypeInfo
				}
			}

			fields[field.Name] = fieldInfo
		}
	}

	return fields, nil
}

// resolvedTypeToString converts a ResolvedType to its string representation, returning an error for unknown types that cannot be resolved to a valid string representation. This ensures that callers can handle type resolution failures appropriately instead of receiving invalid strings like "unknown" or "map[unknown]unknown".
func resolvedTypeToString(rt *ResolvedType) (string, error) {
	if rt == nil {
		return "", fmt.Errorf("resolved type is nil")
	}

	if rt.IsPointer && rt.Underlying != nil {
		underlyingStr, err := resolvedTypeToString(rt.Underlying)
		if err != nil {
			return "", fmt.Errorf("failed to stringify pointer underlying type: %w", err)
		}
		return "*" + underlyingStr, nil
	}
	if rt.IsSlice && rt.KeyType != nil {
		elemStr, err := resolvedTypeToString(rt.KeyType)
		if err != nil {
			return "", fmt.Errorf("failed to stringify slice element type: %w", err)
		}
		return "[]" + elemStr, nil
	}
	if rt.IsMap && rt.KeyType != nil && rt.ValueType != nil {
		keyStr, err := resolvedTypeToString(rt.KeyType)
		if err != nil {
			return "", fmt.Errorf("failed to stringify map key type: %w", err)
		}
		valueStr, err := resolvedTypeToString(rt.ValueType)
		if err != nil {
			return "", fmt.Errorf("failed to stringify map value type: %w", err)
		}
		return fmt.Sprintf("map[%s]%s", keyStr, valueStr), nil
	}
	if rt.Package != "" {
		return rt.Package + "." + rt.Name, nil
	}
	if rt.Name != "" {
		return rt.Name, nil
	}
	return "", fmt.Errorf("unable to determine string representation for resolved type with kind %v", rt.Kind)
}

// typeToString converts a ParsedType to its string representation, returning an error for unknown types that cannot be resolved to a valid string representation. This ensures that callers can handle type resolution failures appropriately instead of receiving invalid strings like "unknown" or "map[unknown]unknown".
func typeToString(pt ParsedType) (string, error) {
	if pt.IsPointer {
		elemStr, err := typeToString(*pt.KeyType)
		if err != nil {
			return "", fmt.Errorf("failed to stringify pointer element type: %w", err)
		}
		return "*" + elemStr, nil
	}
	if pt.IsSlice {
		elemStr, err := typeToString(*pt.KeyType)
		if err != nil {
			return "", fmt.Errorf("failed to stringify slice element type: %w", err)
		}
		return "[]" + elemStr, nil
	}
	if pt.IsMap {
		if pt.KeyType == nil || pt.ValueType == nil {
			return "", fmt.Errorf("map type missing key or value type")
		}
		keyStr, err := typeToString(*pt.KeyType)
		if err != nil {
			return "", fmt.Errorf("failed to stringify map key type: %w", err)
		}
		valueStr, err := typeToString(*pt.ValueType)
		if err != nil {
			return "", fmt.Errorf("failed to stringify map value type: %w", err)
		}
		return fmt.Sprintf("map[%s]%s", keyStr, valueStr), nil
	}
	if pt.Package != "" {
		return pt.Package + "." + pt.Name, nil
	}
	if pt.Name != "" {
		return pt.Name, nil
	}
	return "", fmt.Errorf("unable to determine string representation for type with kind %v", pt.Kind)
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
