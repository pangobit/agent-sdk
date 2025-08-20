package tools

import (
	"fmt"
	"reflect"
	"strings"
)

// ServiceRegistry defines the interface for service registration
type ServiceRegistry interface {
	Register(service any) error
}

// JSONRPCMethodExecutor implements MethodExecutor using a service registry
type JSONRPCMethodExecutor struct {
	registry ServiceRegistry
	services map[string]any
}

// NewJSONRPCMethodExecutor creates a new JSON-RPC method executor
func NewJSONRPCMethodExecutor(registry ServiceRegistry) *JSONRPCMethodExecutor {
	return &JSONRPCMethodExecutor{
		registry: registry,
		services: make(map[string]any),
	}
}

// RegisterService registers a service with the registry
func (e *JSONRPCMethodExecutor) RegisterService(service any) error {
	// Register with registry for validation
	if err := e.registry.Register(service); err != nil {
		return err
	}

	// Also store locally for direct access
	serviceType := reflect.TypeOf(service)
	if serviceType.Kind() == reflect.Ptr {
		serviceType = serviceType.Elem()
	}
	serviceName := serviceType.Name()
	e.services[serviceName] = service

	return nil
}

// ExecuteMethod executes a method by directly calling the registered service
func (e *JSONRPCMethodExecutor) ExecuteMethod(serviceName, methodName string, params map[string]interface{}) (interface{}, error) {
	// Get the service
	service, exists := e.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service '%s' not found", serviceName)
	}

	// Get the service value
	serviceValue := reflect.ValueOf(service)

	// Find the method
	method := serviceValue.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method '%s' not found in service '%s'", methodName, serviceName)
	}

	// Get method type
	methodType := method.Type()

	// Check if method has the correct signature (should have 2 parameters: request and response pointer)
	if methodType.NumIn() != 2 {
		return nil, fmt.Errorf("method '%s' must have exactly 2 parameters (request and response pointer)", methodName)
	}

	// Create request parameter
	requestType := methodType.In(0)
	requestValue := reflect.New(requestType).Elem()

	// Convert params map to request struct
	if err := e.mapToStruct(params, requestValue); err != nil {
		return nil, fmt.Errorf("failed to convert parameters to request: %w", err)
	}

	// Create response parameter
	responseType := methodType.In(1)
	responseValue := reflect.New(responseType.Elem())

	// Call the method
	results := method.Call([]reflect.Value{requestValue, responseValue})

	// Check for errors
	if len(results) > 0 && !results[0].IsNil() {
		return nil, results[0].Interface().(error)
	}

	// Return the response
	return responseValue.Elem().Interface(), nil
}

// mapToStruct converts a map to a struct using reflection
func (e *JSONRPCMethodExecutor) mapToStruct(m map[string]interface{}, v reflect.Value) error {
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a struct")
	}

	vType := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := vType.Field(i)

		// Get JSON tag name
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = fieldType.Name
		}
		// Remove comma and options from JSON tag
		if commaIndex := strings.Index(jsonTag, ","); commaIndex != -1 {
			jsonTag = jsonTag[:commaIndex]
		}

		// Get value from map
		if value, exists := m[jsonTag]; exists {
			// Convert and set the field
			if err := e.setFieldValue(field, value); err != nil {
				return fmt.Errorf("failed to set field '%s': %w", fieldType.Name, err)
			}
		}
	}

	return nil
}

// setFieldValue sets a field value with type conversion
func (e *JSONRPCMethodExecutor) setFieldValue(field reflect.Value, value interface{}) error {
	fieldType := field.Type()

	// Handle different types
	switch fieldType.Kind() {
	case reflect.String:
		if str, ok := value.(string); ok {
			field.SetString(str)
		} else {
			return fmt.Errorf("cannot convert %v to string", value)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v := value.(type) {
		case int:
			field.SetInt(int64(v))
		case float64:
			field.SetInt(int64(v))
		default:
			return fmt.Errorf("cannot convert %v to int", value)
		}
	case reflect.Float32, reflect.Float64:
		if f, ok := value.(float64); ok {
			field.SetFloat(f)
		} else {
			return fmt.Errorf("cannot convert %v to float", value)
		}
	case reflect.Bool:
		if b, ok := value.(bool); ok {
			field.SetBool(b)
		} else {
			return fmt.Errorf("cannot convert %v to bool", value)
		}
	default:
		return fmt.Errorf("unsupported field type: %v", fieldType.Kind())
	}

	return nil
}
