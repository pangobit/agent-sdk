package tools

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewToolService(t *testing.T) {
	tests := []struct {
		name string
		opts []ToolServiceOpts
		want *ToolService
	}{
		{
			name: "default_tool_service",
			opts: nil,
			want: &ToolService{
				methodRegistry: make(map[string]MethodInfo),
			},
		},
		{
			name: "tool_service_with_options",
			opts: []ToolServiceOpts{
				func(ts *ToolService) {
					ts.methodRegistry["test"] = MethodInfo{
						ServiceName: "TestService",
						MethodName:  "TestMethod",
						Description: "Test description",
						Parameters:  map[string]interface{}{},
					}
				},
			},
			want: &ToolService{
				methodRegistry: map[string]MethodInfo{
					"test": {
						ServiceName: "TestService",
						MethodName:  "TestMethod",
						Description: "Test description",
						Parameters:  map[string]interface{}{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewToolService(tt.opts...)

			// Check that the service was created
			if got == nil {
				t.Fatal("NewToolService returned nil")
			}

			// Check that the method registry was initialized
			if got.methodRegistry == nil {
				t.Error("methodRegistry was not initialized")
			}

			// For the test with options, verify the registry was populated
			if len(tt.opts) > 0 {
				registry := got.GetMethodRegistry()
				if len(registry) != len(tt.want.methodRegistry) {
					t.Errorf("expected %d methods in registry, got %d",
						len(tt.want.methodRegistry), len(registry))
				}
			}
		})
	}
}

func TestToolService_RegisterMethod(t *testing.T) {
	tests := []struct {
		name          string
		serviceName   string
		methodName    string
		description   string
		parameters    map[string]interface{}
		expectedKey   string
		expectedError bool
	}{
		{
			name:          "register_valid_method",
			serviceName:   "HelloService",
			methodName:    "Hello",
			description:   "Sends a greeting",
			parameters:    map[string]interface{}{"name": "string"},
			expectedKey:   "HelloService.Hello",
			expectedError: false,
		},
		{
			name:          "register_method_with_empty_strings",
			serviceName:   "",
			methodName:    "",
			description:   "Empty method",
			parameters:    map[string]interface{}{},
			expectedKey:   ".",
			expectedError: false,
		},
		{
			name:          "register_method_with_nil_parameters",
			serviceName:   "TestService",
			methodName:    "TestMethod",
			description:   "Test method",
			parameters:    nil,
			expectedKey:   "TestService.TestMethod",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewToolService()

			err := ts.RegisterMethod(tt.serviceName, tt.methodName, tt.description, tt.parameters)

			if tt.expectedError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Verify the method was registered
			registry := ts.GetMethodRegistry()
			methodInfo, exists := registry[tt.expectedKey]

			if !exists {
				t.Errorf("method %s was not registered", tt.expectedKey)
				return
			}

			if methodInfo.ServiceName != tt.serviceName {
				t.Errorf("expected service name %s, got %s", tt.serviceName, methodInfo.ServiceName)
			}
			if methodInfo.MethodName != tt.methodName {
				t.Errorf("expected method name %s, got %s", tt.methodName, methodInfo.MethodName)
			}
			if methodInfo.Description != tt.description {
				t.Errorf("expected description %s, got %s", tt.description, methodInfo.Description)
			}
			if !reflect.DeepEqual(methodInfo.Parameters, tt.parameters) {
				t.Errorf("expected parameters %v, got %v", tt.parameters, methodInfo.Parameters)
			}
		})
	}
}

func TestToolService_GetMethodRegistry(t *testing.T) {
	tests := []struct {
		name            string
		registerMethods []struct {
			serviceName, methodName, description string
			parameters                           map[string]interface{}
		}
		expectedCount int
	}{
		{
			name:            "empty_registry",
			registerMethods: nil,
			expectedCount:   0,
		},
		{
			name: "single_method",
			registerMethods: []struct {
				serviceName, methodName, description string
				parameters                           map[string]interface{}
			}{
				{"HelloService", "Hello", "Greeting method", map[string]interface{}{"name": "string"}},
			},
			expectedCount: 1,
		},
		{
			name: "multiple_methods",
			registerMethods: []struct {
				serviceName, methodName, description string
				parameters                           map[string]interface{}
			}{
				{"HelloService", "Hello", "Greeting method", map[string]interface{}{"name": "string"}},
				{"UserService", "CreateUser", "Create user method", map[string]interface{}{"email": "string"}},
				{"UserService", "GetUser", "Get user method", map[string]interface{}{"id": "int"}},
			},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewToolService()

			// Register methods
			for _, method := range tt.registerMethods {
				err := ts.RegisterMethod(method.serviceName, method.methodName, method.description, method.parameters)
				if err != nil {
					t.Fatalf("failed to register method: %v", err)
				}
			}

			// Get registry
			registry := ts.GetMethodRegistry()

			if len(registry) != tt.expectedCount {
				t.Errorf("expected %d methods in registry, got %d", tt.expectedCount, len(registry))
			}

			// Verify it's a copy (not the original map)
			if len(tt.registerMethods) > 0 {
				// Modify the returned registry
				for key := range registry {
					delete(registry, key)
				}

				// Get registry again and verify it still has the original data
				registry2 := ts.GetMethodRegistry()
				if len(registry2) != tt.expectedCount {
					t.Error("registry was not properly copied - modifications affected original")
				}
			}
		})
	}
}

func TestToolService_ToolDiscoveryHandler(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		registerMethods []struct {
			serviceName, methodName, description string
			parameters                           map[string]interface{}
		}
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:            "get_tools_with_no_methods",
			method:          http.MethodGet,
			registerMethods: nil,
			expectedStatus:  http.StatusOK,
			expectedBody: map[string]interface{}{
				"tools":       map[string]interface{}{},
				"description": "Available tools for LLM-powered applications",
			},
		},
		{
			name:   "get_tools_with_methods",
			method: http.MethodGet,
			registerMethods: []struct {
				serviceName, methodName, description string
				parameters                           map[string]interface{}
			}{
				{"HelloService", "Hello", "Greeting method", map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name to greet",
						"required":    true,
					},
				}},
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"tools": map[string]interface{}{
					"HelloService.Hello": map[string]interface{}{
						"name":        "HelloService.Hello",
						"description": "Greeting method",
						"parameters": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"description": "Name to greet",
								"required":    true,
							},
						},
						"returns": "",
					},
				},
				"description": "Available tools for LLM-powered applications",
			},
		},
		{
			name:            "post_method_not_allowed",
			method:          http.MethodPost,
			registerMethods: nil,
			expectedStatus:  http.StatusMethodNotAllowed,
			expectedBody:    nil,
		},
		{
			name:            "put_method_not_allowed",
			method:          http.MethodPut,
			registerMethods: nil,
			expectedStatus:  http.StatusMethodNotAllowed,
			expectedBody:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewToolService()

			// Register methods
			for _, method := range tt.registerMethods {
				err := ts.RegisterMethod(method.serviceName, method.methodName, method.description, method.parameters)
				if err != nil {
					t.Fatalf("failed to register method: %v", err)
				}
			}

			// Create request
			req := httptest.NewRequest(tt.method, "/tools", nil)
			w := httptest.NewRecorder()

			// Call handler
			handler := ts.ToolDiscoveryHandler()
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

				// Check response structure
				if _, exists := response["tools"]; !exists {
					t.Error("response missing 'tools' field")
				}
				if _, exists := response["description"]; !exists {
					t.Error("response missing 'description' field")
				}

				// For specific test cases, verify the exact content
				if tt.expectedBody != nil {
					// Compare tools count
					expectedTools := tt.expectedBody["tools"].(map[string]interface{})
					actualTools := response["tools"].(map[string]interface{})

					if len(expectedTools) != len(actualTools) {
						t.Errorf("expected %d tools, got %d", len(expectedTools), len(actualTools))
					}

					// For the specific test case, verify tool details
					if len(tt.registerMethods) > 0 {
						expectedKey := tt.registerMethods[0].serviceName + "." + tt.registerMethods[0].methodName
						expectedTool := expectedTools[expectedKey].(map[string]interface{})
						actualTool := actualTools[expectedKey].(map[string]interface{})

						if expectedTool["name"] != actualTool["name"] {
							t.Errorf("expected tool name %s, got %s", expectedTool["name"], actualTool["name"])
						}
						if expectedTool["description"] != actualTool["description"] {
							t.Errorf("expected tool description %s, got %s", expectedTool["description"], actualTool["description"])
						}
					}
				}
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		key      string
		expected interface{}
		testFunc func(map[string]any, string) interface{}
	}{
		{
			name: "getString_valid",
			input: map[string]any{
				"test": "value",
			},
			key:      "test",
			expected: "value",
			testFunc: func(m map[string]any, k string) interface{} { return getString(m, k) },
		},
		{
			name: "getString_missing",
			input: map[string]any{
				"other": "value",
			},
			key:      "test",
			expected: "",
			testFunc: func(m map[string]any, k string) interface{} { return getString(m, k) },
		},
		{
			name: "getString_wrong_type",
			input: map[string]any{
				"test": 123,
			},
			key:      "test",
			expected: "",
			testFunc: func(m map[string]any, k string) interface{} { return getString(m, k) },
		},
		{
			name: "getBool_true",
			input: map[string]any{
				"test": true,
			},
			key:      "test",
			expected: true,
			testFunc: func(m map[string]any, k string) interface{} { return getBool(m, k) },
		},
		{
			name: "getBool_false",
			input: map[string]any{
				"test": false,
			},
			key:      "test",
			expected: false,
			testFunc: func(m map[string]any, k string) interface{} { return getBool(m, k) },
		},
		{
			name: "getBool_missing",
			input: map[string]any{
				"other": true,
			},
			key:      "test",
			expected: false,
			testFunc: func(m map[string]any, k string) interface{} { return getBool(m, k) },
		},
		{
			name: "getBool_wrong_type",
			input: map[string]any{
				"test": "true",
			},
			key:      "test",
			expected: false,
			testFunc: func(m map[string]any, k string) interface{} { return getBool(m, k) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.testFunc(tt.input, tt.key)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
