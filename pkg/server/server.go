package server

import (
	"encoding/json"
	"io"
)

type Transport interface {
	ListenAndServe(addr string) error
}

type Connection interface {
	io.ReadWriteCloser
}

// ToolRegistry defines the interface for tool registration
type ToolRegistry interface {
	RegisterMethod(serviceName, methodName, description string, parameters map[string]interface{}) error
}

type Server struct {
	transport    Transport
	toolRegistry ToolRegistry
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

// RegisterService registers a service with the underlying transport if it supports service registration
func (s *Server) RegisterService(service any) error {
	if hybridTransport, ok := s.transport.(interface{ RegisterWithSchema(interface{}) error }); ok {
		return hybridTransport.RegisterWithSchema(service)
	}
	return nil
}

// RegisterMethod registers a method as a tool
func (s *Server) RegisterMethod(serviceName, methodName, description string, parameters map[string]interface{}) error {
	if s.toolRegistry != nil {
		return s.toolRegistry.RegisterMethod(serviceName, methodName, description, parameters)
	}
	return nil
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

func (s *Server) HandleRequest(path string, reader io.Reader, writer io.Writer) error {
	dec := json.NewDecoder(reader)
	var req map[string]any
	if err := dec.Decode(&req); err != nil {
		return err
	}

	return nil
}

type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Schema      struct {
		Inputs   map[string]any `json:"inputs,omitempty"`
		Required []string       `json:"required,omitempty"`
	} `json:"schema"`
}

type PromptTemplate struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Arguments   map[string]any `json:"arguments"`
}
