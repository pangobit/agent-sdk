# API Generator Web Server Example

This example demonstrates how the `apigen` library can automatically generate API documentation from HTTP handlers in a web server.

## What It Demonstrates

### HTTP Parameter Exclusion
The example shows how `apigen` automatically excludes HTTP-related parameters (`http.ResponseWriter`, `*http.Request`) from the API description, focusing only on the business logic parameters that client libraries need to know about.

### Handler Analysis
The example includes three typical HTTP handlers:

1. **`GetUserProfile`** - Takes a custom `UserProfileRequest` struct
2. **`UpdateUserProfile`** - Takes a custom `UpdateProfileRequest` struct  
3. **`SearchUsers`** - Takes individual parameters: `query string`, `limit int`, `filters map[string]interface{}`

### Generated Output
Running this example shows how `apigen` extracts:

- **Method descriptions** from Go doc comments
- **Parameter names and types** (excluding HTTP types)
- **Structured JSON output** suitable for client libraries

## Key Features Highlighted

- ✅ **Automatic HTTP parameter filtering** - No manual exclusion needed
- ✅ **Custom struct support** - Handles complex request types
- ✅ **Multiple parameter patterns** - Works with structs, individual params, and maps
- ✅ **Documentation preservation** - Maintains Go doc comments in output
- ✅ **JSON serialization** - Output is ready for client consumption

## Usage

```bash
go run examples/apigen_web_server/main.go
```

This will generate and display the API description JSON for the handler methods.

## Integration Pattern

This example shows a common pattern where:

1. **HTTP handlers** are defined with standard Go HTTP signatures
2. **`apigen`** analyzes the handlers to extract API specifications
3. **Client libraries** can consume the generated JSON to understand available endpoints and parameters
4. **Documentation** stays synchronized with code automatically

Perfect for maintaining API documentation in web services with minimal manual effort!