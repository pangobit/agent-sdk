package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/pangobit/agent-sdk/pkg/apigen"
)

func main() {
	// Example: Generate API description from the current package using prefix filter
	config := apigen.WithPrefix("New").SetAPIName("ExampleAPI")

	desc, err := apigen.GenerateFromPackage("./pkg/server/tools", config)
	if err != nil {
		log.Fatalf("Failed to generate API description: %v", err)
	}

	// Pretty print the JSON output
	jsonData, err := json.MarshalIndent(desc, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal to JSON: %v", err)
	}

	fmt.Println("Generated API Description:")
	fmt.Println(string(jsonData))
}