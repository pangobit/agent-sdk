package hybrid

// ToolDescription represents a simple, LLM-friendly tool description
type ToolDescription struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Parameters  map[string]ParameterInfo `json:"parameters"`
	Returns     string                   `json:"returns"`
}

// ParameterBuilder provides a fluent API for building parameter descriptions
type ParameterBuilder struct {
	params map[string]ParameterInfo
}

// NewParameterBuilder creates a new parameter builder
func NewParameterBuilder() *ParameterBuilder {
	return &ParameterBuilder{
		params: make(map[string]ParameterInfo),
	}
}

// String adds a string parameter
func (pb *ParameterBuilder) String(name, description string) *ParameterBuilder {
	pb.params[name] = ParameterInfo{
		Type:        "string",
		Description: description,
		Required:    true,
	}
	return pb
}

// OptionalString adds an optional string parameter
func (pb *ParameterBuilder) OptionalString(name, description string) *ParameterBuilder {
	pb.params[name] = ParameterInfo{
		Type:        "string",
		Description: description,
		Required:    false,
	}
	return pb
}

// Number adds a number parameter
func (pb *ParameterBuilder) Number(name, description string) *ParameterBuilder {
	pb.params[name] = ParameterInfo{
		Type:        "number",
		Description: description,
		Required:    true,
	}
	return pb
}

// OptionalNumber adds an optional number parameter
func (pb *ParameterBuilder) OptionalNumber(name, description string) *ParameterBuilder {
	pb.params[name] = ParameterInfo{
		Type:        "number",
		Description: description,
		Required:    false,
	}
	return pb
}

// Boolean adds a boolean parameter
func (pb *ParameterBuilder) Boolean(name, description string) *ParameterBuilder {
	pb.params[name] = ParameterInfo{
		Type:        "boolean",
		Description: description,
		Required:    true,
	}
	return pb
}

// OptionalBoolean adds an optional boolean parameter
func (pb *ParameterBuilder) OptionalBoolean(name, description string) *ParameterBuilder {
	pb.params[name] = ParameterInfo{
		Type:        "boolean",
		Description: description,
		Required:    false,
	}
	return pb
}

// Array adds an array parameter
func (pb *ParameterBuilder) Array(name, description string) *ParameterBuilder {
	pb.params[name] = ParameterInfo{
		Type:        "array",
		Description: description,
		Required:    true,
	}
	return pb
}

// OptionalArray adds an optional array parameter
func (pb *ParameterBuilder) OptionalArray(name, description string) *ParameterBuilder {
	pb.params[name] = ParameterInfo{
		Type:        "array",
		Description: description,
		Required:    false,
	}
	return pb
}

// Object adds an object parameter
func (pb *ParameterBuilder) Object(name, description string) *ParameterBuilder {
	pb.params[name] = ParameterInfo{
		Type:        "object",
		Description: description,
		Required:    true,
	}
	return pb
}

// OptionalObject adds an optional object parameter
func (pb *ParameterBuilder) OptionalObject(name, description string) *ParameterBuilder {
	pb.params[name] = ParameterInfo{
		Type:        "object",
		Description: description,
		Required:    false,
	}
	return pb
}

// Custom adds a custom parameter with specific type
func (pb *ParameterBuilder) Custom(name, paramType, description string, required bool) *ParameterBuilder {
	pb.params[name] = ParameterInfo{
		Type:        paramType,
		Description: description,
		Required:    required,
	}
	return pb
}

// Build returns the built parameter map
func (pb *ParameterBuilder) Build() map[string]ParameterInfo {
	return pb.params
}

// ToolBuilder provides a fluent API for building tool descriptions
type ToolBuilder struct {
	registry *ToolRegistry
}

// NewToolBuilder creates a new tool builder
func NewToolBuilder(registry *ToolRegistry) *ToolBuilder {
	return &ToolBuilder{
		registry: registry,
	}
}

// Tool registers a tool with the given name and description
func (tb *ToolBuilder) Tool(name, description string) *ToolBuilder {
	tb.registry.RegisterTool(name, description, make(map[string]ParameterInfo))
	return tb
}

// WithParameters adds parameters to the last registered tool
func (tb *ToolBuilder) WithParameters(parameters map[string]ParameterInfo) *ToolBuilder {
	// This would need to be implemented with a more sophisticated approach
	// For now, we'll use the direct registry methods
	return tb
}

// Method registers a method as a tool
func (tb *ToolBuilder) Method(serviceName, methodName, description string, parameters map[string]ParameterInfo) *ToolBuilder {
	tb.registry.RegisterMethod(serviceName, methodName, description, parameters)
	return tb
}
