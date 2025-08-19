package hybrid

import (
	"sync"
)

// ToolRegistry
type ToolRegistry struct {
	tools map[string]ToolInfo
	mutex sync.RWMutex
}

// NewToolRegistry basic constructor
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]ToolInfo),
	}
}

// RegisterTool registers a tool with its description and parameter schema
func (r *ToolRegistry) RegisterTool(name, description string, parameters map[string]ParameterInfo) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.tools[name] = ToolInfo{
		Name:        name,
		Description: description,
		Parameters:  parameters,
	}
}

// RegisterMethod registers a method as a tool with its description and parameter schema
func (r *ToolRegistry) RegisterMethod(serviceName, methodName, description string, parameters map[string]ParameterInfo) {
	toolName := serviceName + "." + methodName
	r.RegisterTool(toolName, description, parameters)
}

// GetTool retrieves a tool by name
func (r *ToolRegistry) GetTool(name string) (ToolInfo, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tool, exists := r.tools[name]
	return tool, exists
}

// GetAllTools returns all registered tools
func (r *ToolRegistry) GetAllTools() map[string]ToolInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Return a copy to avoid race conditions
	tools := make(map[string]ToolInfo)
	for name, tool := range r.tools {
		tools[name] = tool
	}
	return tools
}

// GetToolsList returns tools in LLM-friendly format
func (r *ToolRegistry) GetToolsList() map[string]ToolDescription {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tools := make(map[string]ToolDescription)
	for name, tool := range r.tools {
		tools[name] = ToolDescription{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
			Returns:     tool.Returns,
		}
	}
	return tools
}
