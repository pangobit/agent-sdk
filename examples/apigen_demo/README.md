# API Generation Demo

This example demonstrates the new map output format for the `apigen` package, which generates a `map[string]string` instead of a single JSON string.

## Overview

The traditional apigen output generates a single JSON string containing the entire API description:

```json
{
  "apiName": "MyAPI",
  "methods": {
    "Method1": {"description": "...", "parameters": {...}},
    "Method2": {"description": "...", "parameters": {...}}
  }
}
```

The new map format generates individual JSON strings for each method:

```go
var APIDefinitions = map[string]string{
    "Method1": `{"description": "...", "parameters": {...}}`,
    "Method2": `{"description": "...", "parameters": {...}}`,
}
```

## Benefits

- **Individual Access**: Easy access to method definitions at runtime without parsing the entire API
- **No apiName Field**: Each method JSON contains only `description` and `parameters`
- **Better Performance**: No need to unmarshal the entire API structure to get a single method

## Running the Demo

```bash
go run main.go
```

This will:
1. Show the traditional JSON output format
2. Demonstrate the new map format
3. Show how to access individual method definitions
4. Compare the two approaches

## Using the CLI

### Traditional Format
```bash
go run github.com/pangobit/agent-sdk/cmd/apigen \
  -package=./pkg/server/tools \
  -prefix=New \
  -out=api_gen.go \
  -const=APIJSON
```

### New Map Format
```bash
go run github.com/pangobit/agent-sdk/cmd/apigen \
  -package=./pkg/server/tools \
  -prefix=New \
  -out=api_gen.go \
  -const=APIDefinitions \
  -map
```

## Generated Output

The traditional format creates:
```go
const APIJSON = `{"apiName": "MyAPI", "methods": {...}}`
```

The map format creates:
```go
var APIDefinitions = map[string]string{
    "NewToolService": `{"description": "...", "parameters": {...}}`,
    "NewJSONRPCMethodExecutor": `{"description": "...", "parameters": {...}}`,
}
```

## Usage in Code

```go
// Traditional approach - parse entire API
var api APIDescription
json.Unmarshal([]byte(APIJSON), &api)
method := api.Methods["NewToolService"]

// Map approach - direct access
methodJSON := APIDefinitions["NewToolService"]
var method MethodDescription
json.Unmarshal([]byte(methodJSON), &method)
```</content>
<parameter name="file_path">examples/apigen_demo/README.md