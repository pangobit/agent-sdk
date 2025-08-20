package agentsdk

import (
	"fmt"
	"time"

	"github.com/pangobit/agent-sdk/pkg/jsonrpc"
	"github.com/pangobit/agent-sdk/pkg/server"
	"github.com/pangobit/agent-sdk/pkg/server/http"
	"github.com/pangobit/agent-sdk/pkg/server/tools"
)

// NewDefaultServer creates a new server with the default HTTP transport and tool functionality
func NewDefaultServer() *server.Server {
	// Create tool service for method registration
	toolService := tools.NewToolService()

	// Create JSON-RPC server for service registry
	jsonrpcServer := jsonrpc.NewServer()

	// Create method executor for method execution
	methodExecutor := tools.NewJSONRPCMethodExecutor(jsonrpcServer)

	// Create method execution handler
	methodHandler := tools.NewMethodExecutionHandler(methodExecutor)

	// Create HTTP transport with tool handler and method handler
	httpOpts := []http.HTTPTransportOpts{
		http.WithPath("/agents/api/v1/"),
		http.WithReadDeadline(10 * time.Second),
		http.WithWriteDeadline(10 * time.Second),
		http.WithToolHandler(toolService.ToolDiscoveryHandler()),
		http.WithMethodHandler(methodHandler),
	}
	httpTransport := http.NewHTTPTransport(httpOpts...)

	// Create server with HTTP transport, tool registry, and method executor
	serverOpts := []server.ServerOpts{
		server.WithTransport(httpTransport),
		server.WithToolRegistry(toolService),
		server.WithMethodExecutor(methodExecutor),
	}
	return server.NewServer(serverOpts...)
}

// NewServer creates a new server with HTTP transport
func NewServer(opts ...server.ServerOpts) *server.Server {
	return server.NewServer(opts...)
}

// RegisterService registers a service with the server's method executor.
// A service is a Go struct with methods that can be called via JSON-RPC.
// The service will be available for method execution at the /execute endpoint.
//
// Example:
//
//	type HelloService struct{}
//	func (h *HelloService) Hello(req HelloRequest, reply *HelloResponse) error { ... }
//	agentsdk.RegisterService(server, &HelloService{})
func RegisterService(server *server.Server, service any) error {
	methodExecutor := server.GetMethodExecutor()
	if methodExecutor != nil {
		if registry, ok := methodExecutor.(interface{ RegisterService(any) error }); ok {
			return registry.RegisterService(service)
		}
	}
	return fmt.Errorf("no method executor configured")
}

// DescribeServiceMethod creates a tool description for a service method.
// This allows clients to discover what methods are available and what parameters they require.
// The description will be available at the /tools endpoint for tool discovery.
//
// Think of it this way:
// - RegisterService: "Here's a Go struct with methods you can call"
// - DescribeServiceMethod: "Here's a description of what this method does and what parameters it needs"
//
// Example:
//
//	params := map[string]any{
//	    "name": map[string]any{
//	        "type": "string",
//	        "description": "The name to greet",
//	        "required": true,
//	    },
//	}
//	agentsdk.DescribeServiceMethod(server, "HelloService", "Hello",
//	    "Sends a greeting message to the specified name", params)
func DescribeServiceMethod(server *server.Server, serviceName, methodName, description string, parameters map[string]any) error {
	return server.RegisterMethod(serviceName, methodName, description, parameters)
}
