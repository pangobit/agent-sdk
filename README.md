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
- Describe those services' methods, which provides calling agents a description of what the methods do, and what parameters they accent.

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