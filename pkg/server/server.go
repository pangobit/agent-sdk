package server

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
)

type Connection interface {
	io.ReadWriteCloser
}

type Server struct {
	tools ToolRepository
}

type ServerOpts func(*Server)

func WithToolRepository(repo ToolRepository) ServerOpts {
	return func(s *Server) {
		s.tools = repo
	}
}

func NewServer(opts ...ServerOpts) *Server {
	s := &Server{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Server) Serve(conn Connection) {
	defer conn.Close()

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
	ID          string                   `json:"id"`
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	Arguments   []PromptTemplateArgument `json:"arguments"`
}

type PromptTemplateArgument struct {
	Name     string `json:"name"`
	Argument any    `json:"argument"`
	Required bool   `json:"required"`
}

func (t *PromptTemplateArgument) String() string {
	rendered := struct {
		Name     string             `json:"name"`
		Type     interface{}        `json:"type"`
		Required bool               `json:"required"`
		Items    *map[string]string `json:"items,omitempty"`
	}{
		Name:     t.Name,
		Required: t.Required,
	}

	if t.Argument == nil {
		rendered.Type = nil
	} else {
		argType := reflect.TypeOf(t.Argument)
		rendered.Type = argType.String()

		if argType.Kind() == reflect.Slice || argType.Kind() == reflect.Map {
			items := map[string]string{
				"type": argType.String(),
			}
			rendered.Items = &items
		}
	}

	renderedJSON, err := json.Marshal(rendered)
	if err != nil {
		return ""
	}
	return string(renderedJSON)
}

func (t *PromptTemplate) Invoke() {}

func (s *Server) HTTPHandler(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.ServeHTTP(w, r)
	})
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.HTTPHandler(http.NewServeMux()))
}
