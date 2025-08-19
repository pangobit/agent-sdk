# Hybrid Transport

The hybrid transport provides HTTP endpoints for tool discovery and execution while internally using JSON-RPC for method calls. This implementation follows clean architecture principles with proper separation of concerns.

## Architecture

- **Transport Layer**: Handles HTTP/RPC communication
- **Tool Registry**: Manages tool registration and metadata
- **Service Registration**: Registers RPC services for method calls

## Usage

### Basic Setup

```go
// Create hybrid transport
transport := hybrid.NewHybridTransport(
    hybrid.WithPath("/api/v1/"),
)

// Register your RPC service
service := &MyService{}
if err := transport.RegisterWithSchema(service); err != nil {
    log.Fatal(err)
}

// Get tool registry for manual tool registration
registry := transport.GetToolRegistry()

// Register tools with semantic descriptions
params := hybrid.NewParameterBuilder().
    String("query", "The search query to execute").
    Number("limit", "Maximum number of results to return").
    OptionalString("filters", "Optional filters to apply").
    Build()

registry.RegisterMethod("MyService", "Search", 
    "Search for items using the provided query and filters", params)
```

### Parameter Builder API

The `ParameterBuilder` provides a fluent API for defining tool parameters:

```go
params := hybrid.NewParameterBuilder().
    String("name", "Full name of the person").
    Number("age", "Age in years").
    Boolean("active", "Whether the person is active").
    OptionalString("email", "Email address").
    Array("tags", "List of tags").
    Object("metadata", "Additional metadata").
    Build()
```

### Available Parameter Types

- `String(name, description)` - Required string parameter
- `OptionalString(name, description)` - Optional string parameter
- `Number(name, description)` - Required number parameter
- `OptionalNumber(name, description)` - Optional number parameter
- `Boolean(name, description)` - Required boolean parameter
- `OptionalBoolean(name, description)` - Optional boolean parameter
- `Array(name, description)` - Required array parameter
- `OptionalArray(name, description)` - Optional array parameter
- `Object(name, description)` - Required object parameter
- `OptionalObject(name, description)` - Optional object parameter
- `Custom(name, type, description, required)` - Custom parameter type

### HTTP Endpoints

- `GET /api/v1/tools` - List all available tools with descriptions
- `POST /api/v1/execute` - Execute a tool (implementation pending)

### RPC Service Requirements

Your RPC service methods must follow this signature:

```go
func (s *MyService) MethodName(req RequestType, resp *ResponseType) error
```

Where:
- `req` is the request parameter (struct or primitive type)
- `resp` is a pointer to the response type
- Returns an error

## Design Principles

1. **Separation of Concerns**: Transport layer is separate from tool management
2. **User-Provided Descriptions**: All tool and parameter descriptions must be provided by the implementor
3. **Clean APIs**: Fluent builder patterns for easy tool registration
4. **Type Safety**: Strong typing for parameter definitions
5. **Semantic Value**: Descriptions are meaningful and LLM-friendly

## Example

See `examples/hybrid/main.go` for a complete working example.
