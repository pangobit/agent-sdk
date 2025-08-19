package hybrid

import (
	"encoding/json"
	"net/http"
)

// ExecuteRequest represents a request to execute a tool
type ExecuteRequest struct {
	Tool       string         `json:"tool"`
	Parameters map[string]any `json:"parameters"`
}

// ExecuteResponse represents the response from tool execution
type ExecuteResponse struct {
	Result any `json:"result"`
}

// toolsHandler handles GET requests for tool discovery
func (t *HybridTransport) toolsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Return LLM-friendly tool descriptions
	tools := t.toolRegistry.GetToolsList()
	response := map[string]interface{}{
		"tools":       tools,
		"description": "Available tools for LLM-powered applications",
	}

	json.NewEncoder(w).Encode(response)
}

// executeHandler handles POST requests for tool execution
func (t *HybridTransport) executeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Placeholder response - will be replaced with actual RPC execution
	response := map[string]interface{}{
		"message": "Execute endpoint - implementation pending",
		"note":    "This will execute tools via internal JSON-RPC",
	}

	json.NewEncoder(w).Encode(response)
}

// handleError handles error responses
func (t *HybridTransport) handleError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	errorResponse := map[string]string{
		"error": err.Error(),
	}

	json.NewEncoder(w).Encode(errorResponse)
}
