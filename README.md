# Agent SDK (In lieu of a clever name)

## Introduction
The term "Agent" has many connotations based on context. In the context of this project, it refers to an application that leverages Large Language Models to perform fully or semi automated tasks.

This library doesn't make assertions about the use case of said Agent, but it is being developed in the context of internal-facing tools, not end user-facing products.

### Purpose
The Agent SDK provides functionality to embed a compliant server within your Go application that exposes tools and resources to Client Agent applications. 


### What about MCP?
[Model Context Protocol](https://modelcontextprotocol.io/) is described in its documentation as being " like a USB-C port for AI applications."
There has been a fairly high level of adoption of MCP across applications, however we have following points of contention with the approach:
* We believe that the specification is unnecessary--such a standard for clients and servers communicating already exists in HTTP. 
* The message framing and specifications for the "primitives" in the spec are verbose and rigid. Ironically, they use semi-structured messages instead of natural language.

Having said that, a secondary (eventual) goal of this library is to serve as a proxy for MCP connections to allow for interoperability with services that have elected the MCP as their standard.

# Getting Started
## Setting up a server
The simplest way to start a server is demonstrated in the ./examples/basic_setup example.
```go
// Create default server
	server := agentsdk.NewDefaultServer()
```
The default server is initialized with some defaults configuration. Worth noting, 
the routes for this server are mounted at `/agents/api/v1/` see "Configuring your server" for more info on how to change this.

Once your server is configured, you need to do two things:
- Register services, aka, types with exported methods that can be called by agents.
- Describe those services' methods, which provides calling agents a description of what the methods do, and what parameters they accept.

### Service Registration Restrictions

Under the hood, service registration uses Go's standard library `net/rpc` package. This imposes specific requirements on the methods that can be registered:

- **Exported Methods**: Only exported methods (starting with capital letters) of exported types can be registered
- **Method Signature**: Methods must have exactly two arguments, both of exported types
- **Pointer Arguments**: The second argument must be a pointer
- **Return Value**: Methods must return exactly one value of type `error`

Clients access registered methods using the format `"Type.Method"`, where `Type` is the receiver's concrete type.

For example, a valid service method would look like:
```go
type MyService struct{}

func (s *MyService) ValidMethod(req MyRequest, resp *MyResponse) error {
    // implementation
    return nil
}
```

### Method Registration: Two Approaches

The Agent SDK provides two ways to describe your service methods, each suited for different use cases:

#### Traditional Structured Registration (`DescribeServiceMethod`)

Use this approach when you want strict type validation and structured parameter schemas. It's ideal for:
- APIs consumed by traditional programmatic clients
- When you need guaranteed type safety and validation
- Building services with well-defined, stable interfaces

```go
// Register with structured parameter schema
parameters := map[string]any{
    "name": map[string]any{
        "type":        "string",
        "description": "The name to greet",
        "required":    true,
    },
    "age": map[string]any{
        "type":        "number",
        "description": "Age in years",
        "minimum":     0,
    },
}

agentsdk.DescribeServiceMethod(server, "UserService", "CreateUser", 
    "Creates a new user account", parameters)
```

**Pros:**
- Strict parameter validation
- Type-safe API contracts
- Well-suited for traditional API clients
- Structured, machine-readable schemas

**Cons:**
- More verbose to write and maintain
- Less flexible for dynamic or complex parameter structures
- Requires careful schema definition

#### LLM-Friendly Registration (`DescribeServiceMethodLLM`)

Use this approach when building APIs that will be consumed by Large Language Models or AI agents. It's ideal for:
- Services designed for AI/LLM consumption
- When you want to provide rich natural language context
- APIs that need flexibility in parameter descriptions
- Building agent-to-agent communication

```go
// Register with combined natural language + structured description
description := `Creates a new user account with the provided information.
Parameters: {"name": {"type": "string", "description": "Full name of the user"}, "email": {"type": "string", "description": "Email address"}, "age": {"type": "number", "description": "Age in years", "minimum": 0}}
The name and email are required. Age is optional and must be non-negative.
Returns: {"userId": "string", "created": "boolean"}`

agentsdk.DescribeServiceMethodLLM(server, "UserService.CreateUser", description, "User creation result")
```

**Pros:**
- More flexible and human-readable descriptions
- Can include examples, edge cases, and natural language guidance
- Better suited for AI/LLM understanding
- Easier to write rich, contextual descriptions

**Cons:**
- Less strict type validation
- More suitable for AI agents than traditional programmatic clients
- Parameter validation relies on description accuracy

#### Choosing Between Approaches

**Use `DescribeServiceMethod` when:**
- Building APIs for traditional software clients
- You need strict type validation and contracts
- Your API has stable, well-defined parameter structures
- You're integrating with existing systems that expect structured schemas

**Use `DescribeServiceMethodLLM` when:**
- Building APIs specifically for AI agents or LLMs
- You want to provide rich contextual information
- Your API parameters are complex or dynamic
- You want flexibility in how parameters are described
- You're building agent-to-agent communication systems

Both approaches produce the same runtime behavior - they just differ in how method information is described and validated.

### Automated API Description Generation with `apigen`

For services with many methods, manually writing descriptions can be time-consuming and error-prone. The `apigen` package provides automated generation of API descriptions from your Go source code.

See [pkg/apigen/README.md](pkg/apigen/README.md) for comprehensive documentation on:
- **Features**: Automatic method discovery, type analysis, doc comment integration
- **Usage**: Library functions and CLI tool for code generation
- **Output Formats**: Traditional JSON and new map-based formats
- **Integration**: Using with `go generate` and embedding in your applications

#### Quick Example

```bash
# Generate map format (recommended for individual method access)
go run github.com/pangobit/agent-sdk/cmd/apigen \
  -package=./pkg/handlers \
  -prefix=Handle \
  -out=api_gen.go \
  -const=APIDefinitions \
  -map
```

See `examples/apigen_demo/` for a demonstration of both output formats.

### Start your server
```go
server.ListenAndServe(":8080")
```

## Configuring your server
The library provides a composable API that leans on the options pattern rather than configuration structs. 
To function as intended, the server requires that you provide a Transport layer, a Tool Registry, and a Method Executor.

The composition done for you in the `agentsdk.NewDefaultServer()` function illustrates how this composition approach works. 

e.g.:
```go
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
```

It is possible to provide your own options methods, or you can use the provided ones in the library. For example, passing `http.WithPath` option to the http.Transport constructor will allow you to override the default base path where the agent sdk endpoints are mounted.

# Contributing
This repo is provided under the MIT license, as a source-available repository. Meaning that, at this time, we're not accepting contributions, so we can't truly call it "open source". That does not mean we will never consider accepting contributions, but it's a question of time and resources.