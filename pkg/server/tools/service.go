// Package tools handles the registration and storage of available tools in the agent server
package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// ToolServiceOpts defines options for configuring the tool service
type ToolServiceOpts func(*ToolService)

// ToolService provides tool registration and discovery capabilities
type ToolService struct {
	// Separate internal storage for each registration mode
	structMethods map[string]structMethodInfo // Key: "ServiceName.MethodName"
	llmMethods    map[string]llmMethodInfo    // Key: "ServiceName.MethodName"
	mutex         sync.RWMutex
}

// structMethodInfo contains data for struct-based method registration
type structMethodInfo struct {
	ServiceName, MethodName, Description string
	Parameters                           map[string]interface{} `json:"omitempty"`
}

// llmMethodInfo contains data for LLM-friendly method registration
type llmMethodInfo struct {
	ServiceName, MethodName string
	ToolDescription         ToolInfoDescription
}

// ToolInfoDescription represents LLM-friendly tool registration
type ToolInfoDescription struct {
	MethodName  string // Full method name (ServiceName.MethodName)
	Description string // Combined natural language + parameter schema for LLM
	Returns     string // Optional: description of return type
}

// ToolInfo represents information about a registered tool for HTTP responses
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Returns     string                 `json:"returns"`
}

// NewToolService creates a new tool service
func NewToolService(opts ...ToolServiceOpts) *ToolService {
	t := &ToolService{
		structMethods: make(map[string]structMethodInfo),
		llmMethods:    make(map[string]llmMethodInfo),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// RegisterMethod registers a method as a tool using struct-based parameters
func (t *ToolService) RegisterMethod(serviceName, methodName, description string, parameters map[string]interface{}) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	methodKey := serviceName + "." + methodName
	t.structMethods[methodKey] = structMethodInfo{
		ServiceName: serviceName,
		MethodName:  methodName,
		Description: description,
		Parameters:  parameters,
	}

	return nil
}

// RegisterMethodLLM registers a method using LLM-friendly combined description
func (t *ToolService) RegisterMethodLLM(methodName, description string, returns ...string) error {
	// Parse method name (format: "ServiceName.MethodName")
	parts := strings.Split(methodName, ".")
	if len(parts) != 2 {
		return fmt.Errorf("method name must be in format 'ServiceName.MethodName'")
	}

	serviceName := parts[0]
	methodNameOnly := parts[1]

	if serviceName == "" || methodNameOnly == "" {
		return fmt.Errorf("service name and method name cannot be empty")
	}

	returnValue := ""
	if len(returns) > 0 {
		returnValue = returns[0]
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.llmMethods[methodName] = llmMethodInfo{
		ServiceName: serviceName,
		MethodName:  methodNameOnly,
		ToolDescription: ToolInfoDescription{
			MethodName:  methodName,
			Description: description,
			Returns:     returnValue,
		},
	}

	return nil
}

// GetMethodRegistry returns a unified view of all registered methods for debugging
func (t *ToolService) GetMethodRegistry() map[string]ToolInfo {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	tools := make(map[string]ToolInfo)

	// Add struct methods
	for key, method := range t.structMethods {
		tools[key] = ToolInfo{
			Name:        key,
			Description: method.Description,
			Parameters:  method.Parameters,
			Returns:     "",
		}
	}

	// Add LLM methods
	for key, method := range t.llmMethods {
		tools[key] = ToolInfo{
			Name:        key,
			Description: method.ToolDescription.Description,
			Parameters:  nil,
			Returns:     method.ToolDescription.Returns,
		}
	}

	return tools
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

		// Get unified view of all methods
		tools := t.GetMethodRegistry()

		response := map[string]interface{}{
			"tools":       tools,
			"description": "Available tools for LLM-powered applications",
		}

		json.NewEncoder(w).Encode(response)
	})
}
