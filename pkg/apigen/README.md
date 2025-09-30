# API Generator Package

The `apigen` package provides functionality to automatically generate JSON-friendly API descriptions from Go source code. It analyzes Go packages or files, extracts method signatures, and generates structured API documentation that can be consumed by client libraries.

## Features

- Parse Go packages or individual files
- Filter methods using prefix, suffix, or contains strategies
- Extract method descriptions from Go doc comments
- Generate parameter information with type details
- Exclude HTTP-related parameters (`http.Request`, `http.ResponseWriter`)
- Output JSON-serializable API descriptions

## Usage

### Basic Example

```go
package main

import (
    "encoding/json"
    "log"

    "github.com/pangobit/agent-sdk/pkg/apigen"
)

func main() {
    // Generate API description for methods starting with "Handle"
    config := apigen.WithPrefix("Handle").SetAPIName("MyAPI")

    desc, err := apigen.GenerateFromPackage("./pkg/myservice", config)
    if err != nil {
        log.Fatalf("Failed to generate API: %v", err)
    }

    // Output as JSON
    jsonData, _ := json.MarshalIndent(desc, "", "  ")
    fmt.Println(string(jsonData))
}
```

### Filtering Strategies

#### Prefix Filter
```go
config := apigen.WithPrefix("Handle")
```

#### Suffix Filter
```go
config := apigen.WithSuffix("Handler")
```

#### Contains Filter
```go
config := apigen.WithContains("Process")
```

#### Specific Method List
```go
config := apigen.WithMethodList("Method1", "Method2", "Method3")
```

### Configuration Options

```go
config := apigen.WithPrefix("API").
    SetAPIName("MyServiceAPI").
    SetExcludeHTTP(true)
```

## API Description Format

The generated API description follows this JSON structure:

```json
{
  "apiName": "MyAPI",
  "maps": [
    {
      "methods": {
        "MethodName": {
          "description": "Method description from Go doc",
          "parameters": {
            "param1": {
              "type": "string"
            },
            "param2": {
              "type": "int"
            }
          }
        }
      }
    }
  ]
}
```

## Data Types

### APIDescription
- `APIName`: Name of the generated API
- `Maps`: Array of MapDescription objects

### MapDescription
- `Methods`: Map of method names to MethodDescription objects

### MethodDescription
- `Description`: Method description extracted from Go doc comments
- `Parameters`: Map of parameter names to ParameterInfo objects

### ParameterInfo
- `Type`: Go type string (e.g., "string", "[]int", "map[string]interface{}")
- `Description`: Optional parameter description (future enhancement)

## Integration with Existing Codebase

This package can be used to automatically generate API descriptions for the existing tool registration system:

```go
// Instead of manually describing methods:
server.RegisterMethod("Service", "Method", "Description", params)

// You can auto-generate from code:
desc, _ := apigen.GenerateFromPackage("./pkg/service", apigen.WithPrefix("Handle"))
// Then register all methods from the description
```

## Error Handling

The package returns errors for:
- Invalid package paths
- Parse errors in Go source files
- Type resolution failures

All functions follow Go's error handling conventions, returning `(result, error)` pairs.