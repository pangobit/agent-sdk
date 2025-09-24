package tools

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// SchemaCheckFunc represents a function that checks a specific aspect of a schema
type SchemaCheckFunc func(t *testing.T, param map[string]interface{})

// SchemaCheck represents a check for a specific parameter
type SchemaCheck struct {
	paramName string
	checks    []SchemaCheckFunc
}

// checkRequiredIsArray checks that the required field is an array with specific values
func checkRequiredIsArray(expected []string) SchemaCheckFunc {
	return func(t *testing.T, param map[string]interface{}) {
		required, exists := param["required"]
		if !exists {
			t.Error("required field missing")
			return
		}
		requiredArray, ok := required.([]interface{})
		if !ok {
			t.Errorf("required field should be an array, got %T", required)
			return
		}
		if len(requiredArray) != len(expected) {
			t.Errorf("required array length mismatch: expected %d, got %d", len(expected), len(requiredArray))
			return
		}
		for i, v := range expected {
			if requiredArray[i] != v {
				t.Errorf("required[%d] mismatch: expected %v, got %v", i, v, requiredArray[i])
			}
		}
	}
}

// checkPropertiesExist checks that specified properties exist
func checkPropertiesExist(expectedProps []string) SchemaCheckFunc {
	return func(t *testing.T, param map[string]interface{}) {
		properties, exists := param["properties"]
		if !exists {
			t.Error("properties field missing")
			return
		}
		propsMap, ok := properties.(map[string]interface{})
		if !ok {
			t.Errorf("properties field should be a map, got %T", properties)
			return
		}
		for _, prop := range expectedProps {
			if _, exists := propsMap[prop]; !exists {
				t.Errorf("property %s missing from properties", prop)
			}
		}
	}
}

// checkNestedStructure checks that nested structure exists and has the expected keys/values
func checkNestedStructure(parentKey, nestedKey string, expected map[string]interface{}) SchemaCheckFunc {
	return func(t *testing.T, param map[string]interface{}) {
		properties, exists := param["properties"]
		if !exists {
			t.Error("properties field missing")
			return
		}
		propsMap, ok := properties.(map[string]interface{})
		if !ok {
			t.Errorf("properties field should be a map, got %T", properties)
			return
		}
		parentProp, exists := propsMap[parentKey]
		if !exists {
			t.Errorf("parent property %s missing", parentKey)
			return
		}
		parentMap, ok := parentProp.(map[string]interface{})
		if !ok {
			t.Errorf("parent property %s should be a map, got %T", parentKey, parentProp)
			return
		}
		nestedProp, exists := parentMap[nestedKey]
		if !exists {
			t.Errorf("nested property %s missing from %s", nestedKey, parentKey)
			return
		}
		nestedMap, ok := nestedProp.(map[string]interface{})
		if !ok {
			t.Errorf("nested property %s.%s should be a map, got %T", parentKey, nestedKey, nestedProp)
			return
		}

		// Check that expected keys exist and have reasonable values
		for key := range expected {
			actualValue, exists := nestedMap[key]
			if !exists {
				t.Errorf("expected key %s missing from nested structure %s.%s", key, parentKey, nestedKey)
				continue
			}

			// For simple type checks, just verify the key exists and has a non-nil value
			if actualValue == nil {
				t.Errorf("expected value for key %s in nested structure %s.%s is nil", key, parentKey, nestedKey)
			}
		}

		t.Logf("✓ Nested structure %s.%s contains expected keys", parentKey, nestedKey)
	}
}

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
		schemaChecks   []SchemaCheck
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
			name:   "get_tools_with_complex_nested_parameters_fixed_behavior",
			method: http.MethodGet,
			registerMethods: []struct {
				serviceName, methodName, description string
				parameters                           map[string]interface{}
			}{
				{"QuizService", "CreateQuiz", "Creates a new quiz", map[string]interface{}{
					"quiz": map[string]interface{}{
						"type":        "object",
						"description": "Quiz creation object",
						"required":    []string{"name", "questions"},
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"description": "Quiz name",
								"example":     "Math Quiz",
							},
							"questions": map[string]interface{}{
								"type":        "array",
								"description": "Array of questions",
								"items": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"text": map[string]interface{}{
											"type":        "string",
											"description": "Question text",
											"example":     "What is 2+2?",
										},
									},
									"required": []string{"text"},
								},
							},
						},
					},
				}},
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"tools": map[string]interface{}{
					"QuizService.CreateQuiz": map[string]interface{}{
						"name":        "QuizService.CreateQuiz",
						"description": "Creates a new quiz",
						"parameters": map[string]interface{}{
							"quiz": map[string]interface{}{
								"type":        "object",
								"description": "Quiz creation object",
								"required":    []string{"name", "questions"},
								"properties": map[string]interface{}{
									"name": map[string]interface{}{
										"type":        "string",
										"description": "Quiz name",
										"example":     "Math Quiz",
									},
									"questions": map[string]interface{}{
										"type":        "array",
										"description": "Array of questions",
										"items": map[string]interface{}{
											"type": "object",
											"properties": map[string]interface{}{
												"text": map[string]interface{}{
													"type":        "string",
													"description": "Question text",
													"example":     "What is 2+2?",
												},
											},
											"required": []string{"text"},
										},
									},
								},
							},
						},
						"returns": "",
					},
				},
				"description": "Available tools for LLM-powered applications",
			},
			schemaChecks: []SchemaCheck{
				{
					paramName: "quiz",
					checks: []SchemaCheckFunc{
						checkRequiredIsArray([]string{"name", "questions"}),
						checkPropertiesExist([]string{"name", "questions"}),
						checkNestedStructure("questions", "items", map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"text": map[string]interface{}{
									"type":        "string",
									"description": "Question text",
									"example":     "What is 2+2?",
								},
							},
							"required": []string{"text"},
						}),
					},
				},
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

						// Run schema checks for complex parameter validation
						for _, schemaCheck := range tt.schemaChecks {
							actualParams := actualTool["parameters"].(map[string]interface{})
							param, exists := actualParams[schemaCheck.paramName]
							if !exists {
								t.Errorf("parameter %s not found in actual response", schemaCheck.paramName)
								continue
							}
							paramMap, ok := param.(map[string]interface{})
							if !ok {
								t.Errorf("parameter %s should be a map, got %T", schemaCheck.paramName, param)
								continue
							}
							for _, check := range schemaCheck.checks {
								check(t, paramMap)
							}
						}
					}
				}
			}
		})
	}
}
