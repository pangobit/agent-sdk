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
	"go/ast"
	"go/token"
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
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Fields      map[string]FieldInfo   `json:"fields,omitempty"`
}

// FieldInfo contains information about a struct field
type FieldInfo struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Annotations map[string]string      `json:"annotations,omitempty"`
	Fields      map[string]FieldInfo   `json:"fields,omitempty"`
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
	Name  string
	Type  ParsedType
	Field *ParsedField // for struct parameters
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