package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/pangobit/agent-sdk/pkg/apigen"
)

func main() {
	fmt.Println("=== API Generation Demo ===")

	// Example: Generate API description from the current package using prefix filter
	config := apigen.WithPrefix("New").SetAPIName("ExampleAPI")

	desc, err := apigen.GenerateFromPackage("./pkg/server/tools", config)
	if err != nil {
		log.Fatalf("Failed to generate API description: %v", err)
	}

	// Traditional JSON output format
	fmt.Println("1. Traditional JSON Output Format:")
	jsonData, err := json.MarshalIndent(desc, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal to JSON: %v", err)
	}
	fmt.Println(string(jsonData))
	fmt.Println()

	// Demonstrate using the library's map generation function
	fmt.Println("2. Using the Library's Map Generation:")

	// Create a temporary example file for demonstration
	exampleContent := `package main

// NewExampleService creates a new example service
func NewExampleService(name string, port int) error {
	return nil
}

// NewConfigLoader creates a new config loader
func NewConfigLoader(path string) error {
	return nil
}
`
	err = os.WriteFile("example.go", []byte(exampleContent), 0644)
	if err != nil {
		log.Fatalf("Failed to create example file: %v", err)
	}
	defer os.Remove("example.go")

	// Generate a Go file with map output using the library
	err = apigen.GenerateAndWriteGoFileFromFileAsMap("example.go", "generated_api.go", "GeneratedAPIDefs", "main", config)
	if err != nil {
		log.Fatalf("Failed to generate Go file: %v", err)
	}
	defer os.Remove("generated_api.go")

	// Read and display the generated file
	generatedContent, err := os.ReadFile("generated_api.go")
	if err != nil {
		log.Fatalf("Failed to read generated file: %v", err)
	}

	fmt.Println("Generated Go file content:")
	fmt.Println(string(generatedContent))
	fmt.Println()

	// Demonstrate how the generated map would be used in real code
	fmt.Println("3. How to Use the Generated Map in Your Application:")
	fmt.Println("The generated map allows runtime access to API definitions.")
	fmt.Println("Each method name maps to its JSON description string.")

	// Show the difference
	fmt.Println("4. Comparison:")
	fmt.Printf("Traditional format: Single JSON string with %d methods\n", len(desc.Methods))
	fmt.Printf("Map format: %d individual JSON strings, one per method\n", len(desc.Methods))
	fmt.Printf("Library generates: Ready-to-use Go code with map[string]string\n")
}