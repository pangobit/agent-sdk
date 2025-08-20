package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pangobit/agent-sdk/pkg/server"
)

// MethodExecutionHandler provides HTTP handlers for method execution
type MethodExecutionHandler struct {
	executor server.MethodExecutor
}

// NewMethodExecutionHandler creates a new method execution handler
func NewMethodExecutionHandler(executor server.MethodExecutor) *MethodExecutionHandler {
	return &MethodExecutionHandler{
		executor: executor,
	}
}

// ServeHTTP handles method execution requests
func (h *MethodExecutionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse the JSON-RPC request
	var request map[string]interface{}
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate JSON-RPC 2.0 request
	if err := h.validateRequest(request); err != nil {
		h.sendErrorResponse(w, request, -32600, "Invalid Request", err.Error())
		return
	}

	// Extract method name
	method, ok := request["method"].(string)
	if !ok {
		h.sendErrorResponse(w, request, -32600, "Invalid Request", "method field is required and must be a string")
		return
	}

	// Parse method name (format: "ServiceName.MethodName")
	serviceName, methodName, err := h.parseMethodName(method)
	if err != nil {
		h.sendErrorResponse(w, request, -32601, "Method not found", err.Error())
		return
	}

	// Extract parameters
	params, err := h.extractParams(request)
	if err != nil {
		h.sendErrorResponse(w, request, -32602, "Invalid params", err.Error())
		return
	}

	// Execute the method
	result, err := h.executor.ExecuteMethod(serviceName, methodName, params)
	if err != nil {
		h.sendErrorResponse(w, request, -32603, "Internal error", err.Error())
		return
	}

	// Send success response
	h.sendSuccessResponse(w, request, result)
}

// validateRequest validates a JSON-RPC 2.0 request
func (h *MethodExecutionHandler) validateRequest(request map[string]interface{}) error {
	// Check JSON-RPC version
	if version, ok := request["jsonrpc"].(string); !ok || version != "2.0" {
		return fmt.Errorf("jsonrpc field must be '2.0'")
	}

	// Check method
	if _, ok := request["method"].(string); !ok {
		return fmt.Errorf("method field is required and must be a string")
	}

	// Check ID (optional but recommended)
	if id, exists := request["id"]; exists && id == nil {
		return fmt.Errorf("id field cannot be null")
	}

	return nil
}

// parseMethodName parses a method name in the format "ServiceName.MethodName"
func (h *MethodExecutionHandler) parseMethodName(method string) (string, string, error) {
	parts := strings.Split(method, ".")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("method name must be in format 'ServiceName.MethodName'")
	}

	serviceName := parts[0]
	methodName := parts[1]

	if serviceName == "" || methodName == "" {
		return "", "", fmt.Errorf("service name and method name cannot be empty")
	}

	return serviceName, methodName, nil
}

// extractParams extracts parameters from the JSON-RPC request
func (h *MethodExecutionHandler) extractParams(request map[string]interface{}) (map[string]interface{}, error) {
	params, exists := request["params"]
	if !exists {
		return make(map[string]interface{}), nil
	}

	// Handle array-style params (JSON-RPC 2.0 allows both array and object)
	switch p := params.(type) {
	case []interface{}:
		if len(p) == 0 {
			return make(map[string]interface{}), nil
		}
		// If it's an array with one element, assume it's the params object
		if paramObj, ok := p[0].(map[string]interface{}); ok {
			return paramObj, nil
		}
		return make(map[string]interface{}), nil
	case map[string]interface{}:
		return p, nil
	default:
		return make(map[string]interface{}), nil
	}
}

// sendSuccessResponse sends a JSON-RPC 2.0 success response
func (h *MethodExecutionHandler) sendSuccessResponse(w http.ResponseWriter, request map[string]interface{}, result interface{}) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"result":  result,
		"id":      request["id"],
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// sendErrorResponse sends a JSON-RPC 2.0 error response
func (h *MethodExecutionHandler) sendErrorResponse(w http.ResponseWriter, request map[string]interface{}, code int, message, data string) {
	errorObj := map[string]interface{}{
		"code":    code,
		"message": message,
	}
	if data != "" {
		errorObj["data"] = data
	}

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"error":   errorObj,
		"id":      request["id"],
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // JSON-RPC 2.0 always returns 200 OK
	json.NewEncoder(w).Encode(response)
}
