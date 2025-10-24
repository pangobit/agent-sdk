package main

import (
	"fmt"
	"log"

	"github.com/pangobit/agent-sdk/pkg/apigen"
)

func main() {
	fmt.Println("=== API Generation Demo ===")

	// New v2.0 API: Clean separation of concerns
	parser := apigen.NewParser()
	transformer := apigen.NewTransformer(apigen.NewTypeRegistry())
	generator := apigen.NewJSONGenerator()

	// Parse methods from package
	methods, err := parser.ParsePackage("./pkg/server/tools")
	if err != nil {
		log.Fatalf("Failed to parse package: %v", err)
	}

	// Filter methods by prefix
	filtered := apigen.FilterByPrefix(methods, "New")

	// Transform to enriched methods with type resolution
	enriched, err := transformer.Transform(filtered)
	if err != nil {
		log.Fatalf("Failed to transform methods: %v", err)
	}

	// Create API description
	desc, err := apigen.NewDescription("ExampleAPI", enriched)
	if err != nil {
		log.Fatalf("Failed to create API description: %v", err)
	}

	// Generate JSON output
	content, err := generator.Generate(desc)
	if err != nil {
		log.Fatalf("Failed to generate JSON: %v", err)
	}

	fmt.Println("Generated JSON API description:")
	fmt.Println(content.Content)

	// Alternative: Generate Go constant
	goGenerator := apigen.NewGoConstGenerator("main", "APIDefinition")
	goContent, err := goGenerator.Generate(desc)
	if err != nil {
		log.Fatalf("Failed to generate Go code: %v", err)
	}

	fmt.Println("\nGenerated Go constant:")
	fmt.Println(goContent.Content)
}
