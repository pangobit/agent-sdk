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

	// Create a temporary example file for demonstration with nested structs and slices
	exampleContent := `package main

// Address represents a physical address
type Address struct {
	Street  string ` + "`" + `json:"street"` + "`" + `
	City    string ` + "`" + `json:"city"` + "`" + `
	State   string ` + "`" + `json:"state"` + "`" + `
	ZipCode string ` + "`" + `json:"zipCode"` + "`" + `
}

// Item represents an item in a collection
type Item struct {
	ID       int    ` + "`" + `json:"id"` + "`" + `
	Name     string ` + "`" + `json:"name"` + "`" + `
	Category string ` + "`" + `json:"category"` + "`" + `
}

// User represents a user with nested address
type User struct {
	ID      int     ` + "`" + `json:"id"` + "`" + `
	Name    string  ` + "`" + `json:"name"` + "`" + `
	Email   string  ` + "`" + `json:"email"` + "`" + `
	Address Address ` + "`" + `json:"address"` + "`" + `
}

// Product represents a product with multiple items
type Product struct {
	ID          int    ` + "`" + `json:"id"` + "`" + `
	Name        string ` + "`" + `json:"name"` + "`" + `
	Description string ` + "`" + `json:"description"` + "`" + `
	Tags        []string ` + "`" + `json:"tags"` + "`" + `
}

// Order represents an order with nested user and items
type Order struct {
	ID       int     ` + "`" + `json:"id"` + "`" + `
	User     User    ` + "`" + `json:"user"` + "`" + `
	Items    []Item  ` + "`" + `json:"items"` + "`" + `
	Products []Product ` + "`" + `json:"products"` + "`" + `
	Total    float64 ` + "`" + `json:"total"` + "`" + `
}

// NewExampleService creates a new example service
func NewExampleService(name string, port int) error {
	return nil
}

// ProcessOrder processes an order with nested data
func ProcessOrder(order Order) error {
	return nil
}

// GetUser retrieves a user by ID
func GetUser(userID int) (User, error) {
	return User{}, nil
}
`
	err := os.WriteFile("example.go", []byte(exampleContent), 0o644)
	if err != nil {
		log.Fatalf("Failed to create example file: %v", err)
	}
	defer os.Remove("example.go")

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

	// Demonstrate nested struct and slice introspection
	fmt.Println("2. Demonstrating Nested Struct and Slice Introspection:")

	// Generate API description for our example file
	exampleConfig := apigen.WithMethodList("ProcessOrder").SetAPIName("OrderAPI")
	exampleDesc, err := apigen.GenerateFromFile("example.go", exampleConfig)
	if err != nil {
		log.Fatalf("Failed to generate example API description: %v", err)
	}

	exampleJSON, err := json.MarshalIndent(exampleDesc, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal example to JSON: %v", err)
	}
	fmt.Println("Generated API description with nested structs and slices:")
	fmt.Println(string(exampleJSON))
	fmt.Println()

	// Demonstrate using the library's map generation function
	fmt.Println("3. Using the Library's Map Generation:")

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
	fmt.Println("4. How to Use the Generated Map in Your Application:")
	fmt.Println("The generated map allows runtime access to API definitions.")
	fmt.Println("Each method name maps to its JSON description string.")

	// Show the difference
	fmt.Println("5. Comparison:")
	fmt.Printf("Traditional format: Single JSON string with %d methods\n", len(desc.Methods))
	fmt.Printf("Map format: %d individual JSON strings, one per method\n", len(desc.Methods))
	fmt.Printf("Library generates: Ready-to-use Go code with map[string]string\n")
	fmt.Println()
	fmt.Println("Advanced features:")
	fmt.Println("- Nested struct introspection: Automatically includes fields of nested structs")
	fmt.Println("- Slice type introspection: Introspects element types of slices")
	fmt.Println("- Cross-package support: Works with types defined in the same package")
}
