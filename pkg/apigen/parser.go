package apigen

import (
	"fmt"
	"go/ast"
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

	// Extract parameters (excluding HTTP types if configured)
	if funcDecl.Type.Params != nil {
		for _, param := range funcDecl.Type.Params.List {
			paramInfo, err := extractParameterInfo(param, fset)
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

// extractParameterInfo extracts type information from an AST field
func extractParameterInfo(field *ast.Field, fset *token.FileSet) (ParameterInfo, error) {
	paramType, err := typeToString(field.Type, fset)
	if err != nil {
		return ParameterInfo{}, err
	}

	return ParameterInfo{
		Type: paramType,
	}, nil
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