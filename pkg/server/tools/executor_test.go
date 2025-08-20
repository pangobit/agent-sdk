package tools

import (
	"fmt"
	"reflect"
	"testing"
)

// MockServiceRegistry implements ServiceRegistry for testing
type MockServiceRegistry struct {
	services      map[string]any
	registerError error
}

func NewMockServiceRegistry() *MockServiceRegistry {
	return &MockServiceRegistry{
		services: make(map[string]any),
	}
}

func (m *MockServiceRegistry) Register(service any) error {
	if m.registerError != nil {
		return m.registerError
	}

	if service == nil {
		return fmt.Errorf("cannot register nil service")
	}

	serviceType := reflect.TypeOf(service)
	if serviceType.Kind() == reflect.Ptr {
		serviceType = serviceType.Elem()
	}
	serviceName := serviceType.Name()
	m.services[serviceName] = service
	return nil
}

// TestService is a mock service for testing
type TestService struct{}

type HelloRequest struct {
	Name string `json:"name"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

func (s *TestService) Hello(req HelloRequest, resp *HelloResponse) error {
	resp.Message = "Hello, " + req.Name + "!"
	return nil
}

func (s *TestService) HelloWithError(req HelloRequest, resp *HelloResponse) error {
	return fmt.Errorf("test error")
}

// TestService2 is another mock service for testing
type TestService2 struct{}

type AddRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}

type AddResponse struct {
	Result int `json:"result"`
}

func (s *TestService2) Add(req AddRequest, resp *AddResponse) error {
	resp.Result = req.A + req.B
	return nil
}

func TestNewJSONRPCMethodExecutor(t *testing.T) {
	tests := []struct {
		name     string
		registry ServiceRegistry
		want     *JSONRPCMethodExecutor
	}{
		{
			name:     "create_with_mock_registry",
			registry: NewMockServiceRegistry(),
			want: &JSONRPCMethodExecutor{
				registry: NewMockServiceRegistry(),
				services: make(map[string]any),
			},
		},
		{
			name:     "create_with_nil_registry",
			registry: nil,
			want: &JSONRPCMethodExecutor{
				registry: nil,
				services: make(map[string]any),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewJSONRPCMethodExecutor(tt.registry)

			if got == nil {
				t.Fatal("NewJSONRPCMethodExecutor returned nil")
			}

			// Don't compare registry pointers directly, just check if they're both nil or both non-nil
			if (tt.want.registry == nil) != (got.registry == nil) {
				t.Errorf("expected registry nil: %t, got registry nil: %t", tt.want.registry == nil, got.registry == nil)
			}

			if got.services == nil {
				t.Error("services map was not initialized")
			}

			if len(got.services) != 0 {
				t.Errorf("expected empty services map, got %d items", len(got.services))
			}
		})
	}
}

func TestJSONRPCMethodExecutor_RegisterService(t *testing.T) {
	tests := []struct {
		name          string
		service       any
		registryError error
		expectedError bool
	}{
		{
			name:          "register_valid_service",
			service:       &TestService{},
			registryError: nil,
			expectedError: false,
		},
		{
			name:          "register_service_with_registry_error",
			service:       &TestService{},
			registryError: fmt.Errorf("registration failed"),
			expectedError: true,
		},
		{
			name:          "register_nil_service",
			service:       nil,
			registryError: nil,
			expectedError: true, // Registry will reject nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRegistry := NewMockServiceRegistry()
			mockRegistry.registerError = tt.registryError

			executor := NewJSONRPCMethodExecutor(mockRegistry)

			err := executor.RegisterService(tt.service)

			if tt.expectedError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// If registration was successful, verify service was stored
			if err == nil && tt.service != nil {
				serviceType := reflect.TypeOf(tt.service)
				if serviceType.Kind() == reflect.Ptr {
					serviceType = serviceType.Elem()
				}
				serviceName := serviceType.Name()

				if _, exists := executor.services[serviceName]; !exists {
					t.Errorf("service %s was not stored in executor", serviceName)
				}
			}
		})
	}
}

func TestJSONRPCMethodExecutor_ExecuteMethod(t *testing.T) {
	tests := []struct {
		name           string
		serviceName    string
		methodName     string
		params         map[string]interface{}
		expectedError  bool
		expectedResult interface{}
	}{
		{
			name:        "execute_valid_method",
			serviceName: "TestService",
			methodName:  "Hello",
			params: map[string]interface{}{
				"name": "World",
			},
			expectedError: false,
			expectedResult: HelloResponse{
				Message: "Hello, World!",
			},
		},
		{
			name:           "execute_nonexistent_service",
			serviceName:    "NonexistentService",
			methodName:     "Hello",
			params:         map[string]interface{}{},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name:           "execute_nonexistent_method",
			serviceName:    "TestService",
			methodName:     "NonexistentMethod",
			params:         map[string]interface{}{},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name:        "execute_method_with_error",
			serviceName: "TestService",
			methodName:  "HelloWithError",
			params: map[string]interface{}{
				"name": "World",
			},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name:        "execute_method_with_invalid_params",
			serviceName: "TestService",
			methodName:  "Hello",
			params: map[string]interface{}{
				"invalid_field": "value",
			},
			expectedError: false, // Invalid fields are ignored
			expectedResult: HelloResponse{
				Message: "Hello, !", // Empty name
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRegistry := NewMockServiceRegistry()
			executor := NewJSONRPCMethodExecutor(mockRegistry)

			// Register the test service
			err := executor.RegisterService(&TestService{})
			if err != nil {
				t.Fatalf("failed to register service: %v", err)
			}

			// Execute the method
			result, err := executor.ExecuteMethod(tt.serviceName, tt.methodName, tt.params)

			if tt.expectedError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check result for successful executions
			if !tt.expectedError && err == nil {
				if result == nil {
					t.Error("expected result but got nil")
				} else {
					// For the specific test case, verify the response
					if tt.methodName == "Hello" {
						if response, ok := result.(HelloResponse); ok {
							if response.Message != tt.expectedResult.(HelloResponse).Message {
								t.Errorf("expected message %s, got %s",
									tt.expectedResult.(HelloResponse).Message, response.Message)
							}
						} else {
							t.Error("result is not of type HelloResponse")
						}
					}
				}
			}
		})
	}
}

func TestJSONRPCMethodExecutor_ExecuteMethodWithDifferentTypes(t *testing.T) {
	tests := []struct {
		name          string
		serviceName   string
		methodName    string
		params        map[string]interface{}
		expectedError bool
	}{
		{
			name:        "execute_add_method",
			serviceName: "TestService2",
			methodName:  "Add",
			params: map[string]interface{}{
				"a": 5,
				"b": 3,
			},
			expectedError: false,
		},
		{
			name:        "execute_add_method_with_float_params",
			serviceName: "TestService2",
			methodName:  "Add",
			params: map[string]interface{}{
				"a": 5.0,
				"b": 3.0,
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRegistry := NewMockServiceRegistry()
			executor := NewJSONRPCMethodExecutor(mockRegistry)

			// Register the test service
			err := executor.RegisterService(&TestService2{})
			if err != nil {
				t.Fatalf("failed to register service: %v", err)
			}

			// Execute the method
			result, err := executor.ExecuteMethod(tt.serviceName, tt.methodName, tt.params)

			if tt.expectedError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check result for successful executions
			if !tt.expectedError && err == nil {
				if response, ok := result.(AddResponse); ok {
					// Handle both int and float64 values in params
					var a, b int
					if aVal, ok := tt.params["a"].(int); ok {
						a = aVal
					} else if aVal, ok := tt.params["a"].(float64); ok {
						a = int(aVal)
					} else {
						t.Errorf("unexpected type for param 'a': %T", tt.params["a"])
						return
					}
					if bVal, ok := tt.params["b"].(int); ok {
						b = bVal
					} else if bVal, ok := tt.params["b"].(float64); ok {
						b = int(bVal)
					} else {
						t.Errorf("unexpected type for param 'b': %T", tt.params["b"])
						return
					}
					expectedSum := a + b
					if response.Result != expectedSum {
						t.Errorf("expected result %d, got %d", expectedSum, response.Result)
					}
				} else {
					t.Error("result is not of type AddResponse")
				}
			}
		})
	}
}

func TestJSONRPCMethodExecutor_mapToStruct(t *testing.T) {
	tests := []struct {
		name          string
		input         map[string]interface{}
		target        reflect.Value
		expectedError bool
	}{
		{
			name: "valid_string_field",
			input: map[string]interface{}{
				"name": "test",
			},
			target:        reflect.ValueOf(HelloRequest{}),
			expectedError: false,
		},
		{
			name: "valid_int_field",
			input: map[string]interface{}{
				"a": 5,
				"b": 3,
			},
			target:        reflect.ValueOf(AddRequest{}),
			expectedError: false,
		},
		{
			name: "valid_float_to_int_conversion",
			input: map[string]interface{}{
				"a": 5.0,
				"b": 3.0,
			},
			target:        reflect.ValueOf(AddRequest{}),
			expectedError: false,
		},
		{
			name: "invalid_field_type",
			input: map[string]interface{}{
				"name": 123, // int instead of string
			},
			target:        reflect.ValueOf(HelloRequest{}),
			expectedError: true,
		},
		{
			name: "missing_field",
			input: map[string]interface{}{
				"other": "value",
			},
			target:        reflect.ValueOf(HelloRequest{}),
			expectedError: false, // Missing fields are ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewJSONRPCMethodExecutor(nil)

			// Create a new instance of the target type
			targetValue := reflect.New(tt.target.Type()).Elem()

			err := executor.mapToStruct(tt.input, targetValue)

			if tt.expectedError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// For successful cases, verify the field was set correctly
			if !tt.expectedError && err == nil {
				if tt.name == "valid_string_field" {
					field := targetValue.FieldByName("Name")
					if field.String() != "test" {
						t.Errorf("expected field value 'test', got '%s'", field.String())
					}
				} else if tt.name == "valid_int_field" || tt.name == "valid_float_to_int_conversion" {
					fieldA := targetValue.FieldByName("A")
					fieldB := targetValue.FieldByName("B")
					if fieldA.Int() != 5 || fieldB.Int() != 3 {
						t.Errorf("expected field values 5 and 3, got %d and %d", fieldA.Int(), fieldB.Int())
					}
				}
			}
		})
	}
}

func TestJSONRPCMethodExecutor_setFieldValue(t *testing.T) {
	tests := []struct {
		name          string
		fieldType     reflect.Type
		value         interface{}
		expectedError bool
	}{
		{
			name:          "set_string_field",
			fieldType:     reflect.TypeOf(""),
			value:         "test",
			expectedError: false,
		},
		{
			name:          "set_int_field",
			fieldType:     reflect.TypeOf(0),
			value:         42,
			expectedError: false,
		},
		{
			name:          "set_float_field",
			fieldType:     reflect.TypeOf(0.0),
			value:         3.14,
			expectedError: false,
		},
		{
			name:          "set_bool_field",
			fieldType:     reflect.TypeOf(false),
			value:         true,
			expectedError: false,
		},
		{
			name:          "set_int_from_float",
			fieldType:     reflect.TypeOf(0),
			value:         42.0,
			expectedError: false,
		},
		{
			name:          "set_string_from_int_error",
			fieldType:     reflect.TypeOf(""),
			value:         42,
			expectedError: true,
		},
		{
			name:          "set_int_from_string_error",
			fieldType:     reflect.TypeOf(0),
			value:         "not_a_number",
			expectedError: true,
		},
		{
			name:          "set_unsupported_type",
			fieldType:     reflect.TypeOf([]string{}),
			value:         []string{"test"},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewJSONRPCMethodExecutor(nil)

			// Create a field of the specified type
			fieldValue := reflect.New(tt.fieldType).Elem()

			err := executor.setFieldValue(fieldValue, tt.value)

			if tt.expectedError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// For successful cases, verify the value was set correctly
			if !tt.expectedError && err == nil {
				switch tt.fieldType.Kind() {
				case reflect.String:
					if fieldValue.String() != tt.value.(string) {
						t.Errorf("expected string value %s, got %s", tt.value.(string), fieldValue.String())
					}
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					var expected int64
					switch v := tt.value.(type) {
					case int:
						expected = int64(v)
					case float64:
						expected = int64(v)
					default:
						t.Errorf("unexpected value type: %T", tt.value)
						return
					}
					if fieldValue.Int() != expected {
						t.Errorf("expected int value %d, got %d", expected, fieldValue.Int())
					}
				case reflect.Float32, reflect.Float64:
					if fieldValue.Float() != tt.value.(float64) {
						t.Errorf("expected float value %f, got %f", tt.value.(float64), fieldValue.Float())
					}
				case reflect.Bool:
					if fieldValue.Bool() != tt.value.(bool) {
						t.Errorf("expected bool value %t, got %t", tt.value.(bool), fieldValue.Bool())
					}
				}
			}
		})
	}
}
