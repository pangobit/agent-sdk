package agentsdk

import (
	"time"

	"github.com/pangobit/agent-sdk/pkg/server"
	"github.com/pangobit/agent-sdk/pkg/server/http"
	"github.com/pangobit/agent-sdk/pkg/server/tools"
)

// NewDefaultServer creates a new server with the default HTTP transport and tool functionality
func NewDefaultServer() *server.Server {
	// Create tool service for method registration
	toolService := tools.NewToolService()

	// Create HTTP transport with tool handler
	httpOpts := []http.HTTPTransportOpts{
		http.WithPath("/agents/api/v1/"),
		http.WithReadDeadline(10 * time.Second),
		http.WithWriteDeadline(10 * time.Second),
		http.WithToolHandler(toolService.ToolDiscoveryHandler()),
	}
	httpTransport := http.NewHTTPTransport(httpOpts...)

	// Create server with HTTP transport and tool registry
	serverOpts := []server.ServerOpts{
		server.WithTransport(httpTransport),
		server.WithToolRegistry(toolService),
	}
	return server.NewServer(serverOpts...)
}

// NewServer creates a new server with HTTP transport
func NewServer(opts ...server.ServerOpts) *server.Server {
	return server.NewServer(opts...)
}
