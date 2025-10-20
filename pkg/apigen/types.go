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
	"fmt"
	"go/ast"
	"go/token"
	"sort"
)

// APIDescription represents a complete API description
type APIDescription struct {
	APIName string                       `json:"apiName"`
	Methods map[string]MethodDescription `json:"methods"`
}

// MethodDescription contains information about a method
type MethodDescription struct {
	Description string                 `json:"description"`
	Parameters  map[string]ParameterInfo `json:"parameters"`
}

// ParameterInfo contains information about a parameter
type ParameterInfo struct {
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Fields      map[string]FieldInfo `json:"fields,omitempty"`
	
	// For slices and maps at parameter level
	ElementType *FieldInfo `json:"elementType,omitempty"`
	KeyType     *FieldInfo `json:"keyType,omitempty"`
	ValueType   *FieldInfo `json:"valueType,omitempty"`
}

// FieldInfo contains information about a struct field
type FieldInfo struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Annotations map[string]string      `json:"annotations,omitempty"`
	Fields      map[string]FieldInfo   `json:"fields,omitempty"`
	
	// For slices and maps
	ElementType *FieldInfo `json:"elementType,omitempty"`
	KeyType     *FieldInfo `json:"keyType,omitempty"`
	ValueType   *FieldInfo `json:"valueType,omitempty"`
}

// TypeKind represents the kind of a Go type
type TypeKind int

const (
	TypeKindBasic TypeKind = iota
	TypeKindPointer
	TypeKindSlice
	TypeKindArray
	TypeKindMap
	TypeKindStruct
	TypeKindInterface
	TypeKindFunc
	TypeKindSelector // package.Type
	TypeKindUnknown
)

// RawMethod represents a parsed method from AST
type RawMethod struct {
	Name     string
	Doc      []string
	Params   []RawParam
	Position token.Position
}

// RawParam represents a parsed parameter from AST
type RawParam struct {
	Name string
	Type ast.Expr
}

// ParsedType represents a fully resolved type
type ParsedType struct {
	Kind      TypeKind
	Name      string
	Package   string
	IsPointer bool
	IsSlice   bool
	IsMap     bool
	KeyType   *ParsedType
	ValueType *ParsedType
	Fields    []ParsedField
}

// ParsedField represents a parsed struct field
type ParsedField struct {
	Name string
	Type ParsedType
	Tags map[string]string
}

// EnrichedMethod represents a method with fully resolved types
type EnrichedMethod struct {
	Name        string
	Description string
	Parameters  []EnrichedParam
}

// EnrichedParam represents a parameter with fully resolved type
type EnrichedParam struct {
	Name         string
	Type         ParsedType
	Field        *ParsedField // for struct parameters
	ResolvedType *ResolvedType // resolved type information
}

// ResolvedType represents a fully resolved type with all references resolved
type ResolvedType struct {
	Kind       TypeKind
	Name       string
	Package    string
	IsPointer  bool
	IsSlice    bool
	IsMap      bool
	KeyType    *ResolvedType
	ValueType  *ResolvedType
	Fields     []ResolvedField
	Underlying *ResolvedType // for pointers, slices, etc.
}

// ResolvedField represents a resolved struct field
type ResolvedField struct {
	Name string
	Type *ResolvedType
	Tags map[string]string
}

// TypeResolver handles resolution of type references
type TypeResolver struct {
	registry    *TypeRegistry
	resolved    map[string]*ResolvedType
	resolving   map[string]bool // to detect cycles
	imports     map[string]string // package name -> import path
}

// NewTypeResolver creates a new type resolver
func NewTypeResolver(registry *TypeRegistry) *TypeResolver {
	return &TypeResolver{
		registry:  registry,
		resolved:  make(map[string]*ResolvedType),
		resolving: make(map[string]bool),
		imports:   make(map[string]string),
	}
}

// ResolveType resolves a type by name, handling dependencies
func (r *TypeResolver) ResolveType(name string) (*ResolvedType, error) {
	if resolved, exists := r.resolved[name]; exists {
		return resolved, nil
	}

	if r.resolving[name] {
		return nil, fmt.Errorf("circular type dependency detected for %s", name)
	}

	r.resolving[name] = true
	defer delete(r.resolving, name)

	parsedType, exists := r.registry.GetType(name)
	if !exists {
		// Assume it's a built-in type
		return &ResolvedType{
			Kind: TypeKindBasic,
			Name: name,
		}, nil
	}

	resolved, err := r.resolveParsedType(parsedType)
	if err != nil {
		return nil, err
	}

	r.resolved[name] = resolved
	return resolved, nil
}

// resolveParsedType converts a ParsedType to ResolvedType
func (r *TypeResolver) resolveParsedType(pt *ParsedType) (*ResolvedType, error) {
	rt := &ResolvedType{
		Kind:      pt.Kind,
		Name:      pt.Name,
		Package:   pt.Package,
		IsPointer: pt.IsPointer,
		IsSlice:   pt.IsSlice,
		IsMap:     pt.IsMap,
	}

	// Resolve named types
	if pt.Kind == TypeKindBasic && pt.Name != "" {
		if resolvedType, err := r.ResolveType(pt.Name); err == nil {
			return resolvedType, nil
		}
		// If resolution fails, keep as basic type
	}

	// Resolve nested types
	if pt.KeyType != nil {
		keyType, err := r.resolveParsedType(pt.KeyType)
		if err != nil {
			return nil, err
		}
		rt.KeyType = keyType
	}

	if pt.ValueType != nil {
		valueType, err := r.resolveParsedType(pt.ValueType)
		if err != nil {
			return nil, err
		}
		rt.ValueType = valueType
	}

	// Resolve fields
	for _, field := range pt.Fields {
		fieldType, err := r.resolveParsedType(&field.Type)
		if err != nil {
			return nil, err
		}
		rt.Fields = append(rt.Fields, ResolvedField{
			Name: field.Name,
			Type: fieldType,
			Tags: field.Tags,
		})
	}

	return rt, nil
}

// AddImport adds a package import for cross-package resolution
func (r *TypeResolver) AddImport(packageName, importPath string) {
	r.imports[packageName] = importPath
}

// ResolveAllTypes resolves all types in the registry
func (r *TypeResolver) ResolveAllTypes() error {
	// Get all type names
	var typeNames []string
	for name := range r.registry.types {
		typeNames = append(typeNames, name)
	}

	// Sort for deterministic resolution (simple types first)
	sort.Slice(typeNames, func(i, j int) bool {
		// Try to resolve simpler types first
		ti := r.registry.types[typeNames[i]]
		tj := r.registry.types[typeNames[j]]
		
		// Basic types first
		if ti.Kind != tj.Kind {
			return ti.Kind < tj.Kind
		}
		return typeNames[i] < typeNames[j]
	})

	// Resolve each type
	for _, name := range typeNames {
		_, err := r.ResolveType(name)
		if err != nil {
			return fmt.Errorf("failed to resolve type %s: %w", name, err)
		}
	}

	return nil
}

// TypeRegistry holds type definitions for resolution
type TypeRegistry struct {
	types map[string]*ParsedType
}

// AddType adds a type to the registry
func (r *TypeRegistry) AddType(name string, typ *ParsedType) {
	if r.types == nil {
		r.types = make(map[string]*ParsedType)
	}
	r.types[name] = typ
}

// GetType retrieves a type from the registry
func (r *TypeRegistry) GetType(name string) (*ParsedType, bool) {
	if r.types == nil {
		return nil, false
	}
	typ, exists := r.types[name]
	return typ, exists
}

// resolveType recursively resolves type references
func (r *TypeRegistry) resolveType(pt *ParsedType) {
	// Resolve named types
	if pt.Kind == TypeKindBasic && pt.Name != "" {
		if resolvedType, exists := r.GetType(pt.Name); exists && resolvedType != pt {
			// Copy the resolved type's properties
			*pt = *resolvedType
		}
	}

	// Recursively resolve nested types
	if pt.KeyType != nil {
		r.resolveType(pt.KeyType)
	}
	if pt.ValueType != nil {
		r.resolveType(pt.ValueType)
	}

	// Resolve fields
	for i := range pt.Fields {
		r.resolveType(&pt.Fields[i].Type)
	}
}

// ResolveAllTypes resolves all type references in the registry using the new resolver
func (r *TypeRegistry) ResolveAllTypes() {
	resolver := NewTypeResolver(r)
	resolver.ResolveAllTypes() // Ignore error for backward compatibility
}

// NewTypeRegistry creates a new type registry
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		types: make(map[string]*ParsedType),
	}
}

// GeneratedContent represents generated output
type GeneratedContent struct {
	Content     string
	FileName    string
	ConstName   string
	PackageName string
}

// ParseStrategy defines how to filter methods
type ParseStrategy int

const (
	StrategyPrefix ParseStrategy = iota
	StrategySuffix
	StrategyContains
)

// GeneratorConfig configures the API generation process (legacy)
type GeneratorConfig struct {
	Strategy      ParseStrategy // How to filter methods
	Filter        string        // Filter string (prefix, suffix, or contains)
	MethodList    []string      // Optional discrete list of methods to include
	ExcludeHTTP   bool          // Whether to exclude HTTP-related parameters
	APIName       string        // Name for the generated API
}

// Parser interface
type Parser interface {
	ParsePackage(packagePath string) ([]RawMethod, error)
	ParseSingleFile(filePath string) ([]RawMethod, error)
	GetRegistry() *TypeRegistry
}

// Transformer interface
type Transformer interface {
	Transform(methods []RawMethod) ([]EnrichedMethod, error)
}

// Generator interface for different output formats
type Generator interface {
	Generate(desc APIDescription) (GeneratedContent, error)
}

// Writer interface
type Writer interface {
	WriteToFile(content GeneratedContent, filePath string) error
}