package apigen

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
	Type        string            `json:"type"`
	Description string            `json:"description,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ParseStrategy defines how to filter methods
type ParseStrategy int

const (
	StrategyPrefix ParseStrategy = iota
	StrategySuffix
	StrategyContains
)

// GeneratorConfig configures the API generation process
type GeneratorConfig struct {
	Strategy      ParseStrategy // How to filter methods
	Filter        string        // Filter string (prefix, suffix, or contains)
	MethodList    []string      // Optional discrete list of methods to include
	ExcludeHTTP   bool          // Whether to exclude HTTP-related parameters
	APIName       string        // Name for the generated API
}