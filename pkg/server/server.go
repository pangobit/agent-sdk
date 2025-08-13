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

type Server struct {
	tools     ToolRepository
	transport Transport
}

type ServerOpts func(*Server)

func WithToolRepository(repo ToolRepository) ServerOpts {
	return func(s *Server) {
		s.tools = repo
	}
}

func WithTransport(transport Transport) ServerOpts {
	return func(s *Server) {
		s.transport = transport
	}
}

func NewServer(opts ...ServerOpts) *Server {
	s := &Server{}
	for _, opt := range opts {
		opt(s)
	}
	return s
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
