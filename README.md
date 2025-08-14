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

Having said that, a secondary goal of this library is to serve as a proxy for MCP connections to allow for interoperability with services that have elected the MCP as their standard.

## Getting Started
```
Coming soon
```