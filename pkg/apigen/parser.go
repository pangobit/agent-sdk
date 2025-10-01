package apigen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// extractMethods extracts all function/method declarations from an AST file
func extractMethods(file *ast.File) map[string]*ast.FuncDecl {
	methods := make(map[string]*ast.FuncDecl)

	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			name := funcDecl.Name.Name
			methods[name] = funcDecl
		}
	}

	return methods
}

// filterMethods filters methods based on the configuration strategy
func filterMethods(methods map[string]*ast.FuncDecl, config GeneratorConfig) map[string]*ast.FuncDecl {
	filtered := make(map[string]*ast.FuncDecl)

	// If method list is provided, use it as whitelist
	if len(config.MethodList) > 0 {
		for _, methodName := range config.MethodList {
			if method, exists := methods[methodName]; exists {
				filtered[methodName] = method
			}
		}
		return filtered
	}

	// Otherwise, apply strategy-based filtering
	for name, method := range methods {
		if matchesFilter(name, config) {
			filtered[name] = method
		}
	}

	return filtered
}

// matchesFilter checks if a method name matches the filter criteria
func matchesFilter(methodName string, config GeneratorConfig) bool {
	switch config.Strategy {
	case StrategyPrefix:
		return strings.HasPrefix(methodName, config.Filter)
	case StrategySuffix:
		return strings.HasSuffix(methodName, config.Filter)
	case StrategyContains:
		return strings.Contains(methodName, config.Filter)
	default:
		return false
	}
}

// generateMethodDescription creates a MethodDescription from an AST function declaration
func generateMethodDescription(funcDecl *ast.FuncDecl, fset *token.FileSet) (MethodDescription, error) {
	desc := MethodDescription{
		Parameters: make(map[string]ParameterInfo),
	}

	// Extract description from comments
	desc.Description = extractDocComment(funcDecl)

	// Collect type definitions from the file for struct field resolution
	typeDefs := collectTypeDefinitions(funcDecl, fset)

	// Extract parameters (excluding HTTP types if configured)
	if funcDecl.Type.Params != nil {
		for _, param := range funcDecl.Type.Params.List {
			paramInfo, err := extractParameterInfo(param, fset, typeDefs)
			if err != nil {
				return desc, fmt.Errorf("failed to extract parameter info: %w", err)
			}

			// Skip if it's an HTTP-related type
			if isHTTPType(paramInfo.Type) {
				continue
			}

			// Handle multiple parameters with same type
			for _, name := range param.Names {
				desc.Parameters[name.Name] = paramInfo
			}
		}
	}

	return desc, nil
}

// extractDocComment extracts documentation comment from a function declaration
func extractDocComment(funcDecl *ast.FuncDecl) string {
	if funcDecl.Doc != nil && len(funcDecl.Doc.List) > 0 {
		var comments []string
		for _, comment := range funcDecl.Doc.List {
			text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
			if text != "" {
				comments = append(comments, text)
			}
		}
		return strings.Join(comments, " ")
	}
	return ""
}

// collectTypeDefinitions collects all type definitions from the file containing the function
func collectTypeDefinitions(funcDecl *ast.FuncDecl, fset *token.FileSet) map[string]*ast.TypeSpec {
	typeDefs := make(map[string]*ast.TypeSpec)

	// Find the file that contains this function
	pos := fset.Position(funcDecl.Pos())
	filename := pos.Filename

	// We need to parse the file again to get all declarations
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return typeDefs
	}

	// Collect all type declarations
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					typeDefs[typeSpec.Name.Name] = typeSpec
				}
			}
		}
	}

	return typeDefs
}

// extractParameterInfo extracts type information from an AST field
func extractParameterInfo(field *ast.Field, fset *token.FileSet, typeDefs map[string]*ast.TypeSpec) (ParameterInfo, error) {
	paramType, err := typeToString(field.Type, fset)
	if err != nil {
		return ParameterInfo{}, err
	}

	paramInfo := ParameterInfo{
		Type: paramType,
	}

	// If this is a custom struct type defined in the same file, extract its fields
	if ident, ok := field.Type.(*ast.Ident); ok {
		if typeSpec, exists := typeDefs[ident.Name]; exists {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				fields, err := extractStructFields(structType, fset)
				if err == nil {
					paramInfo.Fields = fields
				}
			}
		}
	}

	return paramInfo, nil
}

// extractStructFields extracts field information from a struct type
func extractStructFields(structType *ast.StructType, fset *token.FileSet) (map[string]FieldInfo, error) {
	fields := make(map[string]FieldInfo)

	if structType.Fields == nil {
		return fields, nil
	}

	for _, field := range structType.Fields.List {
		fieldType, err := typeToString(field.Type, fset)
		if err != nil {
			continue // Skip fields we can't parse
		}

		fieldInfo := FieldInfo{
			Type: fieldType,
		}

		// Parse struct tags if present
		if field.Tag != nil {
			annotations, err := parseStructTags(field.Tag.Value)
			if err == nil {
				fieldInfo.Annotations = annotations
			}
		}

		// Handle multiple field names with same type
		for _, name := range field.Names {
			fields[name.Name] = fieldInfo
		}
	}

	return fields, nil
}

// parseStructTags parses Go struct tags into a map of key-value pairs
// Go struct tags are in the format `key:"value" key2:"value2"`
func parseStructTags(tagStr string) (map[string]string, error) {
	annotations := make(map[string]string)

	// Remove surrounding quotes if present
	tagStr = strings.Trim(tagStr, "`")

	// Simple parser for struct tags
	// This handles the common format: key:"value" key2:"value2,option"
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

// typeToString converts an AST type expression to a string representation
func typeToString(expr ast.Expr, fset *token.FileSet) (string, error) {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name, nil
	case *ast.SelectorExpr:
		pkg := t.X.(*ast.Ident).Name
		return fmt.Sprintf("%s.%s", pkg, t.Sel.Name), nil
	case *ast.StarExpr:
		elemType, err := typeToString(t.X, fset)
		if err != nil {
			return "", err
		}
		return "*" + elemType, nil
	case *ast.ArrayType:
		elemType, err := typeToString(t.Elt, fset)
		if err != nil {
			return "", err
		}
		return "[]" + elemType, nil
	case *ast.MapType:
		keyType, err := typeToString(t.Key, fset)
		if err != nil {
			return "", err
		}
		valueType, err := typeToString(t.Value, fset)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("map[%s]%s", keyType, valueType), nil
	case *ast.InterfaceType:
		return "interface{}", nil
	case *ast.StructType:
		return "struct{}", nil
	case *ast.FuncType:
		return "func", nil
	default:
		// For more complex types, use the token position to get source text
		if fset != nil {
			start := fset.Position(expr.Pos())
			return fmt.Sprintf("<%s>", start.String()), nil
		}
		return "unknown", nil
	}
}

// isHTTPType checks if a type is HTTP-related and should be excluded
func isHTTPType(typeStr string) bool {
	httpTypes := []string{
		"http.Request",
		"*http.Request",
		"http.ResponseWriter",
		"ResponseWriter", // interface type
	}

	for _, httpType := range httpTypes {
		if typeStr == httpType {
			return true
		}
	}

	return false
}