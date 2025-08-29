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
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Returns     string                 `json:"returns"`
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

		// Convert method registry to tool info format with full schema preservation
		tools := make(map[string]ToolInfo)
		registry := t.GetMethodRegistry()

		for methodKey, methodInfo := range registry {
			tools[methodKey] = ToolInfo{
				Name:        methodKey,
				Description: methodInfo.Description,
				Parameters:  methodInfo.Parameters, // Preserve full parameter schema
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
