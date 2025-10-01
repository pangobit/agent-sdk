package main

import (
	"fmt"
	"log"

	agentsdk "github.com/pangobit/agent-sdk/pkg"
)

// Simple math service for demonstration
type MathService struct{}

// AddRequest represents the request for addition
type AddRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}

// AddResponse represents the response for addition
type AddResponse struct {
	Result int `json:"result"`
}

// Add performs addition of two numbers
func (m *MathService) Add(req AddRequest, resp *AddResponse) error {
	resp.Result = req.A + req.B
	return nil
}

// MultiplyRequest represents the request for multiplication
type MultiplyRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}

// MultiplyResponse represents the response for multiplication
type MultiplyResponse struct {
	Result int `json:"result"`
}

// Multiply performs multiplication of two numbers
func (m *MathService) Multiply(req MultiplyRequest, resp *MultiplyResponse) error {
	resp.Result = req.A * req.B
	return nil
}

func main() {
	// Create a new default server
	server := agentsdk.NewDefaultServer()

	// Register the service
	if err := agentsdk.RegisterService(server, &MathService{}); err != nil {
		log.Fatal(err)
	}

	// Register methods using LLM-friendly descriptions
	// These descriptions contain both natural language and structured parameter information
	// that LLMs can understand and use to generate appropriate calls

	addDescription := `Adds two integers together.
Parameters: {"a": {"type": "integer", "description": "First number to add"}, "b": {"type": "integer", "description": "Second number to add"}}
Returns the sum of the two numbers.`

	if err := agentsdk.DescribeServiceMethodLLM(server, "MathService.Add", addDescription, "integer"); err != nil {
		log.Fatal(err)
	}

	multiplyDescription := `Multiplies two integers together.
Parameters: {"a": {"type": "integer", "description": "First number to multiply"}, "b": {"type": "integer", "description": "Second number to multiply"}}
Returns the product of the two numbers.`

	if err := agentsdk.DescribeServiceMethodLLM(server, "MathService.Multiply", multiplyDescription, "integer"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("LLM-Friendly Math Service")
	fmt.Println("========================")
	fmt.Println("Server starting on :8080")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /agents/api/v1/tools   - Tool discovery with LLM descriptions")
	fmt.Println("  POST /agents/api/v1/execute - Method execution")
	fmt.Println("")
	fmt.Println("Example tool discovery:")
	fmt.Println("  curl http://localhost:8080/agents/api/v1/tools")
	fmt.Println("")
	fmt.Println("Example method execution:")
	fmt.Println("  curl -X POST http://localhost:8080/agents/api/v1/execute \\")
	fmt.Println("    -H 'Content-Type: application/json' \\")
	fmt.Println("    -d '{\"jsonrpc\": \"2.0\", \"method\": \"MathService.Add\", \"params\": {\"a\": 5, \"b\": 3}, \"id\": 1}'")

	log.Fatal(server.ListenAndServe(":8080"))
}