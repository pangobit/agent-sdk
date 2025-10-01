## Mixed Registration Example

This example demonstrates both traditional struct-based method registration and the new LLM-friendly combined description registration approach.

### Features Demonstrated

1. **Traditional Struct-Based Registration**: Uses `DescribeServiceMethod` with structured parameter schemas
2. **LLM-Friendly Registration**: Uses `DescribeServiceMethodLLM` with combined natural language + schema descriptions

### Running the Example

```bash
go run main.go
```

The server will start on port 8080. Visit `http://localhost:8080/agents/api/v1/tools` to see the registered tools.

### Tool Discovery Response

The `/tools` endpoint will return information about both registered methods:

```json
{
  "tools": {
    "CalculatorService.Add": {
      "name": "CalculatorService.Add",
      "description": "Adds two integers together",
      "parameters": {
        "a": {"type": "integer", "description": "First number to add", "required": true},
        "b": {"type": "integer", "description": "Second number to add", "required": true}
      },
      "returns": ""
    },
    "CalculatorService.Multiply": {
      "name": "CalculatorService.Multiply", 
      "description": "Multiplies two numbers together.\nParameters: {\"x\": {\"type\": \"number\", \"description\": \"First number\"}, \"y\": {\"type\": \"number\", \"description\": \"Second number\"}}\nReturns: The product of x and y as a number",
      "parameters": null,
      "returns": "number"
    }
  },
  "description": "Available tools for LLM-powered applications"
}
```

### Key Differences

- **Struct Mode** (`CalculatorService.Add`): Uses structured parameter schemas that are validated against Go struct fields
- **LLM Mode** (`CalculatorService.Multiply`): Uses a combined description that includes both natural language and parameter schemas, giving implementors full flexibility in how they describe their APIs for LLM consumption

Both modes work with the same JSON-RPC execution endpoint, but LLM mode provides more flexibility for describing complex or dynamic APIs.