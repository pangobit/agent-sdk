package mcp

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
