package server

import (
	"fmt"
)

type Transport interface {
	ListenAndServe(addr string) error
}

// ToolRegistry defines the interface for tool registration
type ToolRegistry interface {
	RegisterMethod(serviceName, methodName, description string, parameters map[string]interface{}) error
}

// MethodExecutor defines the interface for executing methods
type MethodExecutor interface {
	ExecuteMethod(serviceName, methodName string, params map[string]interface{}) (interface{}, error)
}

type Server struct {
	transport      Transport
	toolRegistry   ToolRegistry
	methodExecutor MethodExecutor
}

type ServerOpts func(*Server)

type TransportOpts func(Transport) Transport

func WithTransport(transport Transport, opts ...TransportOpts) ServerOpts {
	return func(s *Server) {
		s.transport = transport
		for _, opt := range opts {
			s.transport = opt(s.transport)
		}
	}
}

func WithToolRegistry(registry ToolRegistry) ServerOpts {
	return func(s *Server) {
		s.toolRegistry = registry
	}
}

func WithMethodExecutor(executor MethodExecutor) ServerOpts {
	return func(s *Server) {
		s.methodExecutor = executor
	}
}

func NewServer(opts ...ServerOpts) *Server {
	s := &Server{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// GetTransport returns the underlying transport
func (s *Server) GetTransport() Transport {
	return s.transport
}

// GetToolRegistry returns the tool registry
func (s *Server) GetToolRegistry() ToolRegistry {
	return s.toolRegistry
}

// GetMethodExecutor returns the method executor
func (s *Server) GetMethodExecutor() MethodExecutor {
	return s.methodExecutor
}

// RegisterService registers a service with the underlying transport if it supports service registration
func (s *Server) RegisterService(service any) error {
	if hybridTransport, ok := s.transport.(interface{ RegisterWithSchema(interface{}) error }); ok {
		return hybridTransport.RegisterWithSchema(service)
	}
	return nil
}

// RegisterMethod registers a method as a tool
func (s *Server) RegisterMethod(serviceName, methodName, description string, parameters map[string]any) error {
	if s.toolRegistry == nil {
		return fmt.Errorf("tool registry not configured")
	}
	return s.toolRegistry.RegisterMethod(serviceName, methodName, description, parameters)
}

// ExecuteMethod executes a method through the method executor
func (s *Server) ExecuteMethod(serviceName, methodName string, params map[string]any) (any, error) {
	if s.methodExecutor != nil {
		return s.methodExecutor.ExecuteMethod(serviceName, methodName, params)
	}
	return nil, fmt.Errorf("no method executor configured")
}

// HTTPHandler returns the HTTP handler if the transport supports it
func (s *Server) HTTPHandler() any {
	if httpTransport, ok := s.transport.(interface{ HTTPHandler() any }); ok {
		return httpTransport.HTTPHandler()
	}
	return nil
}

func (s *Server) ListenAndServe(addr string) error {
	return s.transport.ListenAndServe(addr)
}
