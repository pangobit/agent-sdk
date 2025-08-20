package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pangobit/agent-sdk/pkg/server"
)

// MockMethodExecutor implements server.MethodExecutor for testing
type MockMethodExecutor struct {
	executeError    error
	executeResult   interface{}
	executeCalled   bool
	lastServiceName string
	lastMethodName  string
	lastParams      map[string]interface{}
}

func NewMockMethodExecutor() *MockMethodExecutor {
	return &MockMethodExecutor{
		executeResult: map[string]interface{}{"result": "success"},
	}
}

func (m *MockMethodExecutor) ExecuteMethod(serviceName, methodName string, params map[string]interface{}) (interface{}, error) {
	m.executeCalled = true
	m.lastServiceName = serviceName
	m.lastMethodName = methodName
	m.lastParams = params

	if m.executeError != nil {
		return nil, m.executeError
	}
	return m.executeResult, nil
}

func TestNewMethodExecutionHandler(t *testing.T) {
	tests := []struct {
		name     string
		executor server.MethodExecutor
		want     *MethodExecutionHandler
	}{
		{
			name:     "create_with_mock_executor",
			executor: NewMockMethodExecutor(),
			want: &MethodExecutionHandler{
				executor: NewMockMethodExecutor(),
			},
		},
		{
			name:     "create_with_nil_executor",
			executor: nil,
			want: &MethodExecutionHandler{
				executor: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMethodExecutionHandler(tt.executor)

			if got == nil {
				t.Fatal("NewMethodExecutionHandler returned nil")
			}

			// Don't compare executor pointers directly, just check if they're both nil or both non-nil
			if (tt.want.executor == nil) != (got.executor == nil) {
				t.Errorf("expected executor nil: %t, got executor nil: %t", tt.want.executor == nil, got.executor == nil)
			}
		})
	}
}

func TestMethodExecutionHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    map[string]interface{}
		executorError  error
		executorResult interface{}
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "valid_jsonrpc_request",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"params":  map[string]interface{}{"name": "World"},
				"id":      1,
			},
			executorError:  nil,
			executorResult: map[string]interface{}{"message": "Hello, World!"},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"result":  map[string]interface{}{"message": "Hello, World!"},
				"id":      float64(1), // JSON numbers are float64
			},
		},
		{
			name:   "get_method_not_allowed",
			method: http.MethodGet,
			requestBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"params":  map[string]interface{}{"name": "World"},
				"id":      1,
			},
			executorError:  nil,
			executorResult: nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
		},
		{
			name:   "put_method_not_allowed",
			method: http.MethodPut,
			requestBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"params":  map[string]interface{}{"name": "World"},
				"id":      1,
			},
			executorError:  nil,
			executorResult: nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
		},
		{
			name:           "invalid_json",
			method:         http.MethodPost,
			requestBody:    nil, // Will send invalid JSON
			executorError:  nil,
			executorResult: nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
		},
		{
			name:   "missing_jsonrpc_version",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"method": "HelloService.Hello",
				"params": map[string]interface{}{"name": "World"},
				"id":     1,
			},
			executorError:  nil,
			executorResult: nil,
			expectedStatus: http.StatusOK, // JSON-RPC 2.0 always returns 200
			expectedBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"error": map[string]interface{}{
					"code":    float64(-32600),
					"message": "Invalid Request",
					"data":    "jsonrpc field must be '2.0'",
				},
				"id": float64(1),
			},
		},
		{
			name:   "wrong_jsonrpc_version",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"jsonrpc": "1.0",
				"method":  "HelloService.Hello",
				"params":  map[string]interface{}{"name": "World"},
				"id":      1,
			},
			executorError:  nil,
			executorResult: nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"error": map[string]interface{}{
					"code":    float64(-32600),
					"message": "Invalid Request",
					"data":    "jsonrpc field must be '2.0'",
				},
				"id": float64(1),
			},
		},
		{
			name:   "missing_method",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"params":  map[string]interface{}{"name": "World"},
				"id":      1,
			},
			executorError:  nil,
			executorResult: nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"error": map[string]interface{}{
					"code":    float64(-32600),
					"message": "Invalid Request",
					"data":    "method field is required and must be a string",
				},
				"id": float64(1),
			},
		},
		{
			name:   "invalid_method_format",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "InvalidMethod",
				"params":  map[string]interface{}{"name": "World"},
				"id":      1,
			},
			executorError:  nil,
			executorResult: nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"error": map[string]interface{}{
					"code":    float64(-32601),
					"message": "Method not found",
					"data":    "method name must be in format 'ServiceName.MethodName'",
				},
				"id": float64(1),
			},
		},
		{
			name:   "empty_service_name",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  ".Hello",
				"params":  map[string]interface{}{"name": "World"},
				"id":      1,
			},
			executorError:  nil,
			executorResult: nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"error": map[string]interface{}{
					"code":    float64(-32601),
					"message": "Method not found",
					"data":    "service name and method name cannot be empty",
				},
				"id": float64(1),
			},
		},
		{
			name:   "empty_method_name",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.",
				"params":  map[string]interface{}{"name": "World"},
				"id":      1,
			},
			executorError:  nil,
			executorResult: nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"error": map[string]interface{}{
					"code":    float64(-32601),
					"message": "Method not found",
					"data":    "service name and method name cannot be empty",
				},
				"id": float64(1),
			},
		},
		{
			name:   "executor_error",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"params":  map[string]interface{}{"name": "World"},
				"id":      1,
			},
			executorError:  fmt.Errorf("service not found"),
			executorResult: nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"error": map[string]interface{}{
					"code":    float64(-32603),
					"message": "Internal error",
					"data":    "service not found",
				},
				"id": float64(1),
			},
		},
		{
			name:   "null_id",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"params":  map[string]interface{}{"name": "World"},
				"id":      nil,
			},
			executorError:  nil,
			executorResult: nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"jsonrpc": "2.0",
				"error": map[string]interface{}{
					"code":    float64(-32600),
					"message": "Invalid Request",
					"data":    "id field cannot be null",
				},
				"id": nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := NewMockMethodExecutor()
			mockExecutor.executeError = tt.executorError
			mockExecutor.executeResult = tt.executorResult

			handler := NewMethodExecutionHandler(mockExecutor)

			// Create request body
			var body []byte
			var err error
			if tt.requestBody != nil {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
			} else {
				body = []byte("invalid json")
			}

			// Create request
			req := httptest.NewRequest(tt.method, "/execute", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Call handler
			handler.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check content type for successful requests
			if tt.expectedStatus == http.StatusOK {
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("expected content type application/json, got %s", contentType)
				}

				// Parse response body
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				// Check JSON-RPC version
				if version, ok := response["jsonrpc"].(string); !ok || version != "2.0" {
					t.Error("response missing or invalid jsonrpc version")
				}

				// Check ID
				if _, exists := response["id"]; !exists {
					t.Error("response missing id field")
				}

				// For specific test cases, verify the exact content
				if tt.expectedBody != nil {
					// Check result or error
					if tt.executorError != nil || tt.requestBody != nil && (tt.requestBody["jsonrpc"] != "2.0" || tt.requestBody["method"] == nil || tt.requestBody["method"] == "InvalidMethod" || tt.requestBody["method"] == ".Hello" || tt.requestBody["method"] == "HelloService." || tt.requestBody["id"] == nil) {
						if _, exists := response["error"]; !exists {
							t.Error("expected error in response but not found")
						}
						expectedError := tt.expectedBody["error"].(map[string]interface{})
						actualError := response["error"].(map[string]interface{})

						if expectedError["code"] != actualError["code"] {
							t.Errorf("expected error code %v, got %v", expectedError["code"], actualError["code"])
						}
						if expectedError["message"] != actualError["message"] {
							t.Errorf("expected error message %v, got %v", expectedError["message"], actualError["message"])
						}
					} else {
						if _, exists := response["result"]; !exists {
							t.Error("expected result in response but not found")
						}
					}
				}
			}

			// Verify executor was called for valid requests
			if tt.expectedStatus == http.StatusOK && tt.executorError == nil && tt.requestBody != nil && tt.requestBody["jsonrpc"] == "2.0" && tt.requestBody["method"] != nil && tt.requestBody["method"] != "InvalidMethod" && tt.requestBody["method"] != ".Hello" && tt.requestBody["method"] != "HelloService." && tt.requestBody["id"] != nil {
				if !mockExecutor.executeCalled {
					t.Error("executor was not called")
				} else {
					// Verify the correct parameters were passed
					if mockExecutor.lastServiceName != "HelloService" {
						t.Errorf("expected service name HelloService, got %s", mockExecutor.lastServiceName)
					}
					if mockExecutor.lastMethodName != "Hello" {
						t.Errorf("expected method name Hello, got %s", mockExecutor.lastMethodName)
					}
					if mockExecutor.lastParams == nil {
						t.Error("params were not passed to executor")
					}
				}
			}
		})
	}
}

func TestMethodExecutionHandler_validateRequest(t *testing.T) {
	tests := []struct {
		name          string
		request       map[string]interface{}
		expectedError bool
		errorMessage  string
	}{
		{
			name: "valid_request",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"params":  map[string]interface{}{"name": "World"},
				"id":      1,
			},
			expectedError: false,
		},
		{
			name: "missing_jsonrpc",
			request: map[string]interface{}{
				"method": "HelloService.Hello",
				"params": map[string]interface{}{"name": "World"},
				"id":     1,
			},
			expectedError: true,
			errorMessage:  "jsonrpc field must be '2.0'",
		},
		{
			name: "wrong_jsonrpc_version",
			request: map[string]interface{}{
				"jsonrpc": "1.0",
				"method":  "HelloService.Hello",
				"params":  map[string]interface{}{"name": "World"},
				"id":      1,
			},
			expectedError: true,
			errorMessage:  "jsonrpc field must be '2.0'",
		},
		{
			name: "missing_method",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"params":  map[string]interface{}{"name": "World"},
				"id":      1,
			},
			expectedError: true,
			errorMessage:  "method field is required and must be a string",
		},
		{
			name: "method_not_string",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  123,
				"params":  map[string]interface{}{"name": "World"},
				"id":      1,
			},
			expectedError: true,
			errorMessage:  "method field is required and must be a string",
		},
		{
			name: "null_id",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"params":  map[string]interface{}{"name": "World"},
				"id":      nil,
			},
			expectedError: true,
			errorMessage:  "id field cannot be null",
		},
		{
			name: "no_id_field",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"params":  map[string]interface{}{"name": "World"},
			},
			expectedError: false, // ID is optional
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewMethodExecutionHandler(nil)

			err := handler.validateRequest(tt.request)

			if tt.expectedError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectedError && err != nil && tt.errorMessage != "" {
				if err.Error() != tt.errorMessage {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			}
		})
	}
}

func TestMethodExecutionHandler_parseMethodName(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		expectedService string
		expectedMethod  string
		expectedError   bool
		errorMessage    string
	}{
		{
			name:            "valid_method_name",
			method:          "HelloService.Hello",
			expectedService: "HelloService",
			expectedMethod:  "Hello",
			expectedError:   false,
		},
		{
			name:            "single_part",
			method:          "HelloService",
			expectedService: "",
			expectedMethod:  "",
			expectedError:   true,
			errorMessage:    "method name must be in format 'ServiceName.MethodName'",
		},
		{
			name:            "three_parts",
			method:          "HelloService.Hello.Extra",
			expectedService: "",
			expectedMethod:  "",
			expectedError:   true,
			errorMessage:    "method name must be in format 'ServiceName.MethodName'",
		},
		{
			name:            "empty_service_name",
			method:          ".Hello",
			expectedService: "",
			expectedMethod:  "",
			expectedError:   true,
			errorMessage:    "service name and method name cannot be empty",
		},
		{
			name:            "empty_method_name",
			method:          "HelloService.",
			expectedService: "",
			expectedMethod:  "",
			expectedError:   true,
			errorMessage:    "service name and method name cannot be empty",
		},
		{
			name:            "both_empty",
			method:          ".",
			expectedService: "",
			expectedMethod:  "",
			expectedError:   true,
			errorMessage:    "service name and method name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewMethodExecutionHandler(nil)

			serviceName, methodName, err := handler.parseMethodName(tt.method)

			if tt.expectedError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectedError && err == nil {
				if serviceName != tt.expectedService {
					t.Errorf("expected service name %s, got %s", tt.expectedService, serviceName)
				}
				if methodName != tt.expectedMethod {
					t.Errorf("expected method name %s, got %s", tt.expectedMethod, methodName)
				}
			}
			if tt.expectedError && err != nil && tt.errorMessage != "" {
				if err.Error() != tt.errorMessage {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			}
		})
	}
}

func TestMethodExecutionHandler_extractParams(t *testing.T) {
	tests := []struct {
		name           string
		request        map[string]interface{}
		expectedParams map[string]interface{}
	}{
		{
			name: "object_params",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"params":  map[string]interface{}{"name": "World"},
				"id":      1,
			},
			expectedParams: map[string]interface{}{"name": "World"},
		},
		{
			name: "array_params_with_object",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"params":  []interface{}{map[string]interface{}{"name": "World"}},
				"id":      1,
			},
			expectedParams: map[string]interface{}{"name": "World"},
		},
		{
			name: "empty_array_params",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"params":  []interface{}{},
				"id":      1,
			},
			expectedParams: map[string]interface{}{},
		},
		{
			name: "array_params_with_non_object",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"params":  []interface{}{"not_an_object"},
				"id":      1,
			},
			expectedParams: map[string]interface{}{},
		},
		{
			name: "no_params_field",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"id":      1,
			},
			expectedParams: map[string]interface{}{},
		},
		{
			name: "null_params",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"params":  nil,
				"id":      1,
			},
			expectedParams: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewMethodExecutionHandler(nil)

			params, err := handler.extractParams(tt.request)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Compare params
			if len(params) != len(tt.expectedParams) {
				t.Errorf("expected %d params, got %d", len(tt.expectedParams), len(params))
			}

			for key, expectedValue := range tt.expectedParams {
				if actualValue, exists := params[key]; !exists {
					t.Errorf("expected param %s not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("expected param %s to be %v, got %v", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestMethodExecutionHandler_sendSuccessResponse(t *testing.T) {
	tests := []struct {
		name     string
		request  map[string]interface{}
		result   interface{}
		expected map[string]interface{}
	}{
		{
			name: "success_response_with_id",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"params":  map[string]interface{}{"name": "World"},
				"id":      1,
			},
			result: map[string]interface{}{"message": "Hello, World!"},
			expected: map[string]interface{}{
				"jsonrpc": "2.0",
				"result":  map[string]interface{}{"message": "Hello, World!"},
				"id":      float64(1),
			},
		},
		{
			name: "success_response_without_id",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"params":  map[string]interface{}{"name": "World"},
			},
			result: "simple result",
			expected: map[string]interface{}{
				"jsonrpc": "2.0",
				"result":  "simple result",
				"id":      nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewMethodExecutionHandler(nil)

			w := httptest.NewRecorder()

			handler.sendSuccessResponse(w, tt.request, tt.result)

			// Check status code
			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			// Check content type
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected content type application/json, got %s", contentType)
			}

			// Parse response
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			// Check JSON-RPC version
			if version, ok := response["jsonrpc"].(string); !ok || version != "2.0" {
				t.Error("response missing or invalid jsonrpc version")
			}

			// Check result
			if _, exists := response["result"]; !exists {
				t.Error("response missing result field")
			}

			// Check ID
			if _, exists := response["id"]; !exists {
				t.Error("response missing id field")
			}

			// Verify no error field
			if _, exists := response["error"]; exists {
				t.Error("success response should not have error field")
			}
		})
	}
}

func TestMethodExecutionHandler_sendErrorResponse(t *testing.T) {
	tests := []struct {
		name          string
		request       map[string]interface{}
		code          int
		message       string
		data          string
		expectedError map[string]interface{}
	}{
		{
			name: "error_response_with_data",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"params":  map[string]interface{}{"name": "World"},
				"id":      1,
			},
			code:    -32603,
			message: "Internal error",
			data:    "service not found",
			expectedError: map[string]interface{}{
				"code":    float64(-32603),
				"message": "Internal error",
				"data":    "service not found",
			},
		},
		{
			name: "error_response_without_data",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "HelloService.Hello",
				"params":  map[string]interface{}{"name": "World"},
				"id":      1,
			},
			code:    -32600,
			message: "Invalid Request",
			data:    "",
			expectedError: map[string]interface{}{
				"code":    float64(-32600),
				"message": "Invalid Request",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewMethodExecutionHandler(nil)

			w := httptest.NewRecorder()

			handler.sendErrorResponse(w, tt.request, tt.code, tt.message, tt.data)

			// Check status code (JSON-RPC 2.0 always returns 200 OK)
			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			// Check content type
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected content type application/json, got %s", contentType)
			}

			// Parse response
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			// Check JSON-RPC version
			if version, ok := response["jsonrpc"].(string); !ok || version != "2.0" {
				t.Error("response missing or invalid jsonrpc version")
			}

			// Check error
			if _, exists := response["error"]; !exists {
				t.Error("response missing error field")
			}

			// Check ID
			if _, exists := response["id"]; !exists {
				t.Error("response missing id field")
			}

			// Verify no result field
			if _, exists := response["result"]; exists {
				t.Error("error response should not have result field")
			}

			// Check error details
			errorObj := response["error"].(map[string]interface{})
			if errorObj["code"] != tt.expectedError["code"] {
				t.Errorf("expected error code %v, got %v", tt.expectedError["code"], errorObj["code"])
			}
			if errorObj["message"] != tt.expectedError["message"] {
				t.Errorf("expected error message %v, got %v", tt.expectedError["message"], errorObj["message"])
			}
			if tt.data != "" {
				if errorObj["data"] != tt.expectedError["data"] {
					t.Errorf("expected error data %v, got %v", tt.expectedError["data"], errorObj["data"])
				}
			} else {
				if _, exists := errorObj["data"]; exists {
					t.Error("error response should not have data field when data is empty")
				}
			}
		})
	}
}
