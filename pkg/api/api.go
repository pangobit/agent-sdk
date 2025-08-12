package api

import (
	"encoding/json"
	"reflect"
)

type Resource string

func GetResources() []Resource {
	return []Resource{
		"resource1",
		"resource2",
		"resource3",
	}
}

type Tool string

func GetTools() []Tool {
	return []Tool{
		"tool1",
		"tool2",
		"tool3",
	}
}

func RegisterTool(tool Tool) {}

func InvokeTool(tool Tool) {}

// Prompt -- We just leverage go templates

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
