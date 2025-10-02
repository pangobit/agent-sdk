package main

import (
	"encoding/json"
	"fmt"
	"log"

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

	// Demonstrate the new map output format
	fmt.Println("2. New Map Output Format:")

	// Generate the map format by iterating through methods
	methodMap := make(map[string]string)
	for methodName, methodDesc := range desc.Methods {
		methodJSON, err := json.MarshalIndent(methodDesc, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal method %s: %v", methodName, err)
		}
		methodMap[methodName] = string(methodJSON)
	}

	fmt.Println("Generated method map:")
	for methodName, methodJSON := range methodMap {
		fmt.Printf("  %s: %s\n", methodName, methodJSON)
	}
	fmt.Println()

	// Show how to access individual methods
	fmt.Println("3. Accessing Individual Methods:")
	if methodJSON, exists := methodMap["NewToolService"]; exists {
		fmt.Printf("Method 'NewToolService' JSON:\n%s\n", methodJSON)
	} else {
		fmt.Println("Method 'NewToolService' not found in generated API")
	}

	fmt.Println("4. Comparison:")
	fmt.Printf("Traditional format: Single JSON string with %d methods\n", len(desc.Methods))
	fmt.Printf("Map format: %d individual JSON strings, one per method\n", len(methodMap))
}