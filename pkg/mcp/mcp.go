package mcp

import (
	"encoding/json"
	"reflect"
)

type client struct {
}

func NewClient() *client {
	return &client{}
}

type server struct {
}

func NewServer() *server {
	return &server{}
}

type ToolCallParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema struct {
		Type       string         `json:"type"`
		Properties map[string]any `json:"properties"`
	} `json:"inputSchema"`
	Required []string `json:"required"`
}

type ResourceTemplate struct {
	URITemplate string `json:"uriTemplate"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	MIMEType    string `json:"mimeType"`
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
