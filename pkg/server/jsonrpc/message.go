package jsonrpc

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// MethodTypeMapping maps method names to their parameter types
type MethodTypeMapping map[string]reflect.Type

// RegisterMethod registers a method with its parameter type
func (m MethodTypeMapping) RegisterMethod(methodName string, paramType reflect.Type) {
	m[methodName] = paramType
}

// GetMethodType retrieves the parameter type for a method
func (m MethodTypeMapping) GetMethodType(methodName string) (reflect.Type, bool) {
	paramType, exists := m[methodName]
	return paramType, exists
}

// MessageProcessor handles message deserialization using Go's type system
type MessageProcessor struct {
	typeMapping *MethodTypeMapping
}

// NewMessageProcessor creates a new message processor
func NewMessageProcessor(typeMapping *MethodTypeMapping) *MessageProcessor {
	return &MessageProcessor{
		typeMapping: typeMapping,
	}
}

// GetTypeMapping returns the method type mapping
func (m *MessageProcessor) GetTypeMapping() *MethodTypeMapping {
	return m.typeMapping
}

// ProcessMessages deserializes messages to the expected Go type for a method
func (m *MessageProcessor) ProcessMessages(methodName string, rawMessages json.RawMessage) (interface{}, error) {
	paramType, exists := m.typeMapping.GetMethodType(methodName)
	if !exists {
		// If no type mapping is found, return the raw messages for backward compatibility
		return rawMessages, nil
	}

	// JSON-RPC sends messages as an array [message], so we need to extract the first element
	var messagesArray [1]any
	if err := json.Unmarshal(rawMessages, &messagesArray); err != nil {
		return nil, fmt.Errorf("invalid message format: %w", err)
	}

	// Extract the first message and marshal it back to JSON for type-safe deserialization
	messageJSON, err := json.Marshal(messagesArray[0])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create a new instance of the expected type
	paramValue := reflect.New(paramType).Interface()

	// Attempt to deserialize the JSON to the expected type
	// This will fail if the JSON structure doesn't match the Go type
	if err := json.Unmarshal(messageJSON, paramValue); err != nil {
		return nil, fmt.Errorf("type mismatch for method '%s': %w", methodName, err)
	}

	// Return the deserialized value (dereference the pointer)
	return reflect.ValueOf(paramValue).Elem().Interface(), nil
}

// GenerateTypeMappingFromStruct generates method type mappings from a Go struct
func GenerateTypeMappingFromStruct(service interface{}) map[string]reflect.Type {
	typeMapping := make(map[string]reflect.Type)

	serviceType := reflect.TypeOf(service)
	if serviceType.Kind() == reflect.Ptr {
		serviceType = serviceType.Elem()
	}

	for i := 0; i < serviceType.NumMethod(); i++ {
		method := serviceType.Method(i)
		if method.Type.NumIn() >= 2 { // Method has parameters (receiver + params)
			paramType := method.Type.In(1)
			typeMapping[method.Name] = paramType
		}
	}

	return typeMapping
}

// MessageFramingService provides message validation and transformation services
type MessageFramingService struct {
	typeMapping *MethodTypeMapping
	processor   *MessageProcessor
}

// NewMessageFramingService creates a new message framing service
func NewMessageFramingService() *MessageFramingService {
	typeMapping := make(MethodTypeMapping)
	processor := NewMessageProcessor(&typeMapping)
	return &MessageFramingService{
		typeMapping: &typeMapping,
		processor:   processor,
	}
}

// RegisterMethodType manually registers a method with its parameter type
func (s *MessageFramingService) RegisterMethodType(methodName string, paramType reflect.Type) {
	s.typeMapping.RegisterMethod(methodName, paramType)
}

// RegisterServiceWithTypeMapping registers a service and generates type mappings for its methods
func (s *MessageFramingService) RegisterServiceWithTypeMapping(service interface{}) error {
	typeMapping := GenerateTypeMappingFromStruct(service)
	for methodName, paramType := range typeMapping {
		s.typeMapping.RegisterMethod(methodName, paramType)
	}
	return nil
}

// ValidateAndTransformMessages validates and transforms messages for a method
func (s *MessageFramingService) ValidateAndTransformMessages(methodName string, rawMessages json.RawMessage) (json.RawMessage, error) {
	result, err := s.processor.ProcessMessages(methodName, rawMessages)
	if err != nil {
		return nil, err
	}

	// If we got a typed result, marshal it back to JSON
	if result != nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		return resultJSON, nil
	}

	// Fallback to raw messages
	return rawMessages, nil
}

// GetTypeMapping returns the method type mapping
func (s *MessageFramingService) GetTypeMapping() *MethodTypeMapping {
	return s.typeMapping
}

// GetProcessor returns the message processor
func (s *MessageFramingService) GetProcessor() *MessageProcessor {
	return s.processor
}
