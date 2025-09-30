package apigen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// GenerateFromPackage generates API description from a Go package
func GenerateFromPackage(packagePath string, config GeneratorConfig) (*APIDescription, error) {
	fset := token.NewFileSet()

	// Parse the package
	pkgs, err := parser.ParseDir(fset, packagePath, func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package: %w", err)
	}

	// Collect all methods from all files
	allMethods := make(map[string]*ast.FuncDecl)

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			methods := extractMethods(file)
			for name, method := range methods {
				allMethods[name] = method
			}
		}
	}

	// Filter methods based on strategy
	filteredMethods := filterMethods(allMethods, config)

	// Generate descriptions
	methods := make(map[string]MethodDescription)

	for name, method := range filteredMethods {
		desc, err := generateMethodDescription(method, fset)
		if err != nil {
			return nil, fmt.Errorf("failed to generate description for %s: %w", name, err)
		}
		methods[name] = desc
	}

	apiDesc := &APIDescription{
		APIName: config.APIName,
		Methods: methods,
	}

	if apiDesc.APIName == "" {
		// Use package name as default
		for pkgName := range pkgs {
			apiDesc.APIName = pkgName
			break
		}
	}

	return apiDesc, nil
}

// GenerateFromFile generates API description from a single Go file
func GenerateFromFile(filePath string, config GeneratorConfig) (*APIDescription, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	methods := extractMethods(file)
	filteredMethods := filterMethods(methods, config)

	// Generate descriptions
	methodDescriptions := make(map[string]MethodDescription)

	for name, method := range filteredMethods {
		desc, err := generateMethodDescription(method, fset)
		if err != nil {
			return nil, fmt.Errorf("failed to generate description for %s: %w", name, err)
		}
		methodDescriptions[name] = desc
	}

	apiDesc := &APIDescription{
		APIName: config.APIName,
		Methods: methodDescriptions,
	}

	if apiDesc.APIName == "" {
		// Use filename (without extension) as default
		base := filepath.Base(filePath)
		apiDesc.APIName = strings.TrimSuffix(base, filepath.Ext(base))
	}

	return apiDesc, nil
}