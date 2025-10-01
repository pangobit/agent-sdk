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
  "methods": {
    "MethodName": {
      "description": "Method description from Go doc",
      "parameters": {
        "param1": {
          "type": "string"
        },
        "customStruct": {
          "type": "CustomRequest",
          "fields": {
            "Field1": {
              "type": "string"
            },
            "Field2": {
              "type": "int"
            }
          }
        }
      }
    }
  }
}
```

### Struct Field Extraction

When a parameter is a custom struct type defined in the same file/package, `apigen` automatically extracts and includes the struct's fields in the API description. This provides complete type information for API consumers.

**Example**: If a method takes a `UserProfileRequest` parameter, the output will include both the type name and all the fields within that struct.

#### Struct Tag Annotations

Struct tags are automatically parsed and included as "annotations" in the field information. Common struct tags include:

- **`json`**: JSON field names and serialization options
- **`bind`**: Parameter binding information (query, path, json, etc.)
- **`validate`**: Validation rules and constraints
- **Custom tags**: Any application-specific metadata

**Example Output:**
```json
"req": {
  "type": "UserProfileRequest",
  "fields": {
    "UserID": {
      "type": "string",
      "annotations": {
        "json": "user_id",
        "bind": "path=user_id,required"
      }
    },
    "IncludeDetails": {
      "type": "bool", 
      "annotations": {
        "json": "include_details",
        "bind": "query=include_details"
      }
    },
    "Format": {
      "type": "string",
      "annotations": {
        "json": "format",
        "bind": "query=format",
        "validate": "omitempty,oneof=json xml"
      }
    }
  }
}
```

## Data Types

### APIDescription
- `APIName`: Name of the generated API
- `Methods`: Map of method names to MethodDescription objects

### MethodDescription
- `Description`: Method description extracted from Go doc comments
- `Parameters`: Map of parameter names to ParameterInfo objects

### ParameterInfo
- `Type`: Go type string (e.g., "string", "[]int", "map[string]interface{}")
- `Fields`: Map of field names to FieldInfo objects (for struct types)
- `Description`: Optional parameter description (future enhancement)

### FieldInfo
- `Type`: Go type string for the struct field
- `Annotations`: Map of struct tag key-value pairs (e.g., json, bind, validate)
- `Description`: Optional field description (future enhancement)

## Go Generate Integration

The `apigen` package provides both library functions and a CLI tool for generating Go files containing API definitions as string constants, making them available at runtime.

### CLI Tool

A command-line tool is available at `cmd/apigen` that can be used with `go generate`:

```bash
go run github.com/pangobit/agent-sdk/cmd/apigen [flags]
```

**Flags:**
- `-package string`: Package path to analyze (cannot use with -file)
- `-file string`: Go file to analyze (cannot use with -package)
- `-out string`: Output Go file path (required)
- `-const string`: Name of the generated constant (required)
- `-api-name string`: Name for the generated API
- `-methods string`: Comma-separated list of method names to include
- `-prefix string`: Include methods starting with this prefix
- `-suffix string`: Include methods ending with this suffix
- `-contains string`: Include methods containing this string

### Using with go:generate

Add a `go:generate` comment to your Go files:

```go
//go:generate go run github.com/pangobit/agent-sdk/cmd/apigen -file=main.go -prefix=Handle -out=api_gen.go -const=APIJSON

func main() {
    // Use the embedded API definition at runtime
    fmt.Println("API:", APIJSON)
}
```

Then run:
```bash
go generate
```

### Library Functions

You can also call the generation functions directly:

```go
config := apigen.WithPrefix("Handle").SetAPIName("MyAPI")

err := apigen.GenerateAndWriteGoFile(
    "./pkg/handlers",     // package path
    "api_gen.go",         // output file
    "MyAPIJSON",          // constant name
    "main",              // package name
    config,
)
```

### Generated File Format

The generated file follows this pattern:

```go
// Code generated by apigen; DO NOT EDIT.
// This file contains the API description for main

package main

// APIJSON contains the JSON API description for this file
const APIJSON = `{
  "apiName": "MyAPI",
  "methods": { ... }
}`
```

This allows you to embed complete API specifications as string constants in your Go binaries!

## Error Handling

The package returns errors for:
- Invalid package paths
- Parse errors in Go source files
- Type resolution failures

All functions follow Go's error handling conventions, returning `(result, error)` pairs.