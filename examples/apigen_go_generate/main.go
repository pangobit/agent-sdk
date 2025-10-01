package main

import (
	"fmt"
)

//go:generate go run github.com/pangobit/agent-sdk/cmd/apigen -file=main.go -methods=ProcessUserData,ValidateInput -out=api_gen.go -const=DataProcessorAPIJSON

func main() {
	// Normal application logic - this will work after running go generate
	fmt.Println("Data Processor API")
	fmt.Println("==================")
	fmt.Println("API Description:", DataProcessorAPIJSON)
}

// UserData represents user information for processing
type UserData struct {
	ID       string `json:"id" validate:"required"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"min=0,max=150"`
	Active   bool   `json:"active"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ProcessUserData processes user data and returns processed result
// Parameters:
//   - userData: The user data to process (required)
func ProcessUserData(userData UserData) (UserData, error) {
	// Simulate processing
	userData.Active = true
	userData.Metadata["processed"] = true
	return userData, nil
}

// ValidationRequest represents a validation request
type ValidationRequest struct {
	Data   interface{} `json:"data" validate:"required"`
	Strict bool        `json:"strict"`
	Rules  []string    `json:"rules"`
}

// ValidateInput validates input data against rules
// Parameters:
//   - request: The validation request containing data and rules
func ValidateInput(request ValidationRequest) (map[string]interface{}, error) {
	// Simulate validation
	result := map[string]interface{}{
		"valid":   true,
		"errors":  []string{},
		"checked": len(request.Rules),
		"strict":  request.Strict,
	}
	return result, nil
}