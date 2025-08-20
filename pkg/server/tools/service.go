package tools

import (
	"encoding/json"
	"net/http"
	"sync"
)

// ToolServiceOpts defines options for configuring the tool service
type ToolServiceOpts func(*ToolService)

// ToolService provides tool registration and discovery capabilities
type ToolService struct {
	methodRegistry map[string]MethodInfo
	mutex          sync.RWMutex
}

// MethodInfo represents information about a registered method
type MethodInfo struct {
	ServiceName string
	MethodName  string
	Description string
	Parameters  map[string]interface{}
}

// ToolInfo represents information about a registered tool for HTTP responses
type ToolInfo struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Parameters  map[string]ParameterInfo `json:"parameters"`
	Returns     string                   `json:"returns"`
}

// ParameterInfo represents information about a tool parameter
type ParameterInfo struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// NewToolService creates a new tool service
func NewToolService(opts ...ToolServiceOpts) *ToolService {
	t := &ToolService{
		methodRegistry: make(map[string]MethodInfo),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// RegisterMethod registers a method as a tool
func (t *ToolService) RegisterMethod(serviceName, methodName, description string, parameters map[string]interface{}) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	methodKey := serviceName + "." + methodName
	t.methodRegistry[methodKey] = MethodInfo{
		ServiceName: serviceName,
		MethodName:  methodName,
		Description: description,
		Parameters:  parameters,
	}

	return nil
}

// GetMethodRegistry returns the method registry
func (t *ToolService) GetMethodRegistry() map[string]MethodInfo {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// Return a copy to avoid race conditions
	registry := make(map[string]MethodInfo)
	for key, info := range t.methodRegistry {
		registry[key] = info
	}
	return registry
}

// ToolDiscoveryHandler returns an HTTP handler for tool discovery
func (t *ToolService) ToolDiscoveryHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Convert method registry to tool info format
		tools := make(map[string]ToolInfo)
		registry := t.GetMethodRegistry()

		for methodKey, methodInfo := range registry {
			// Convert parameters to ParameterInfo format
			paramInfo := make(map[string]ParameterInfo)
			for name, param := range methodInfo.Parameters {
				if paramMap, ok := param.(map[string]interface{}); ok {
					paramInfo[name] = ParameterInfo{
						Type:        getString(paramMap, "type"),
						Description: getString(paramMap, "description"),
						Required:    getBool(paramMap, "required"),
					}
				}
			}

			tools[methodKey] = ToolInfo{
				Name:        methodKey,
				Description: methodInfo.Description,
				Parameters:  paramInfo,
				Returns:     "",
			}
		}

		response := map[string]interface{}{
			"tools":       tools,
			"description": "Available tools for LLM-powered applications",
		}

		json.NewEncoder(w).Encode(response)
	})
}

// Helper functions for type conversion
func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}
