# LLM Description Example

This example demonstrates how to use LLM-friendly method descriptions with the agent SDK. Instead of providing structured parameter schemas, you can provide natural language descriptions that include parameter information in a way that's easily understandable by Large Language Models.

## What It Demonstrates

### LLM-Friendly Method Registration
Instead of using structured parameter maps like this:

```go
parameters := map[string]any{
    "a": map[string]any{
        "type":        "integer",
        "description": "First number to add",
        "required":    true,
    },
    "b": map[string]any{
        "type":        "integer",
        "description": "Second number to add",
        "required":    true,
    },
}
agentsdk.DescribeServiceMethod(server, "MathService", "Add", "Adds two integers", parameters)
```

You can use a combined natural language + structured description:

```go
addDescription := `Adds two integers together.
Parameters: {"a": {"type": "integer", "description": "First number to add"}, "b": {"type": "integer", "description": "Second number to add"}}
Returns the sum of the two numbers.`

agentsdk.DescribeServiceMethodLLM(server, "MathService.Add", addDescription, "integer")
```

### Benefits of LLM Descriptions

1. **More Flexible**: You can mix natural language explanations with structured data
2. **LLM-Optimized**: Designed to be easily parsed and understood by language models
3. **Human-Readable**: Easier to write and maintain than nested map structures
4. **Rich Context**: Can include examples, edge cases, and usage notes

## Running the Example

```bash
go run examples/llm_description_example/main.go
```

The server will start on port 8080. You can:

### View Tool Descriptions
```bash
curl http://localhost:8080/agents/api/v1/tools
```

This will show the LLM-friendly descriptions for both Add and Multiply methods.

### Execute Methods
```bash
# Add two numbers
curl -X POST http://localhost:8080/agents/api/v1/execute \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "MathService.Add", "params": {"a": 5, "b": 3}, "id": 1}'

# Multiply two numbers
curl -X POST http://localhost:8080/agents/api/v1/execute \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "MathService.Multiply", "params": {"a": 4, "b": 7}, "id": 1}'
```

## Key Features

- **Service Registration**: Shows how to register a Go service with methods
- **LLM Method Descriptions**: Demonstrates the new `DescribeServiceMethodLLM` function
- **Combined Descriptions**: Shows how natural language and JSON schemas can coexist
- **Return Type Specification**: Shows how to specify return types for LLM understanding

## When to Use LLM Descriptions

Use `DescribeServiceMethodLLM` when:
- You want more flexibility in describing method parameters
- You're building APIs that will be consumed by LLMs or AI agents
- You want to include rich context, examples, or natural language explanations
- The structured parameter format feels too restrictive

Use traditional `DescribeServiceMethod` when:
- You need strict type validation
- You're building APIs for traditional programmatic clients
- You prefer the structured parameter schema approach