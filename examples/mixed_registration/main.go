package main

import (
	"fmt"
	"log"
	"net/http"

	agentsdk "github.com/pangobit/agent-sdk/pkg"
)

// CalculatorService provides calculation operations
type CalculatorService struct{}

// AddRequest represents the request for addition
type AddRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}

// AddResponse represents the response for addition
type AddResponse struct {
	Result int `json:"result"`
}

func (c *CalculatorService) Add(req AddRequest, resp *AddResponse) error {
	resp.Result = req.A + req.B
	return nil
}

func main() {
	// Create a new default server
	srv := agentsdk.NewDefaultServer()

	// Register the service
	if err := agentsdk.RegisterService(srv, &CalculatorService{}); err != nil {
		log.Fatal(err)
	}

	// Register method using traditional struct-based approach
	parameters := map[string]interface{}{
		"a": map[string]interface{}{
			"type":        "integer",
			"description": "First number to add",
			"required":    true,
		},
		"b": map[string]interface{}{
			"type":        "integer",
			"description": "Second number to add",
			"required":    true,
		},
	}

	if err := agentsdk.DescribeServiceMethod(srv, "CalculatorService", "Add",
		"Adds two integers together", parameters); err != nil {
		log.Fatal(err)
	}

	// Register method using new LLM-friendly approach
	llmDescription := `Multiplies two numbers together.
Parameters: {"x": {"type": "number", "description": "First number"}, "y": {"type": "number", "description": "Second number"}}
Returns: The product of x and y as a number`

	if err := agentsdk.DescribeServiceMethodLLM(srv, "CalculatorService.Multiply",
		llmDescription, "number"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Server starting on :8080")
	fmt.Println("Visit http://localhost:8080/agents/api/v1/tools to see registered tools")

	log.Fatal(http.ListenAndServe(":8080", nil))
}