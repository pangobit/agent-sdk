package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pangobit/agent-sdk/pkg/apigen"
)

func main() {
	fmt.Println("=== Demo Output: Comprehensive API Generation Examples ===")

	// Create output directory for demo files
	outputDir := "demo_outputs"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Initialize the API generation components
	parser := apigen.NewParser()
	transformer := apigen.NewTransformer(parser.GetRegistry())
	jsonGenerator := apigen.NewJSONGenerator()

	// Parse the demo package
	fmt.Println("\n1. Parsing api package...")
	methods, err := parser.ParsePackage("./api")
	if err != nil {
		log.Fatalf("Failed to parse package: %v", err)
	}
	fmt.Printf("   Found %d methods\n", len(methods))

	// Debug: check registry
	registry := parser.GetRegistry()
	fmt.Printf("   Registry has types:")
	if _, exists := registry.GetType("Company"); exists {
		fmt.Printf(" Company")
	}
	if _, exists := registry.GetType("Address"); exists {
		fmt.Printf(" Address")
	}
	if _, exists := registry.GetType("User"); exists {
		fmt.Printf(" User")
	}
	if _, exists := registry.GetType("Team"); exists {
		fmt.Printf(" Team")
	}
	fmt.Println()

	// Demonstrate different filtering strategies
	filters := []struct {
		name   string
		filter func([]apigen.RawMethod) []apigen.RawMethod
	}{
		{"all_methods", func(m []apigen.RawMethod) []apigen.RawMethod { return m }},
		{"handle_prefix", func(m []apigen.RawMethod) []apigen.RawMethod { return apigen.FilterByPrefix(m, "Handle") }},
		{"process_prefix", func(m []apigen.RawMethod) []apigen.RawMethod { return apigen.FilterByPrefix(m, "Process") }},
		{"update_prefix", func(m []apigen.RawMethod) []apigen.RawMethod { return apigen.FilterByPrefix(m, "Update") }},
		{"schedule_prefix", func(m []apigen.RawMethod) []apigen.RawMethod { return apigen.FilterByPrefix(m, "Schedule") }},
	}

	for _, f := range filters {
		fmt.Printf("\n2. Generating demo file for filter: %s\n", f.name)

		filtered := f.filter(methods)
		if len(filtered) == 0 {
			fmt.Printf("   No methods found for filter %s, skipping\n", f.name)
			continue
		}

	// Transform methods
		enriched, err := transformer.Transform(filtered)
		if err != nil {
			log.Fatalf("Failed to transform methods for %s: %v", f.name, err)
		}

		// Debug: check enriched methods
		if f.name == "process_prefix" && len(enriched) > 0 {
			fmt.Printf("   Debug: first method %s has %d params\n", enriched[0].Name, len(enriched[0].Parameters))
			if len(enriched[0].Parameters) > 0 {
				param := enriched[0].Parameters[0]
				fmt.Printf("   Debug: param %s parsed type %s, kind %v, fields %d\n", param.Name, param.Type.Name, param.Type.Kind, len(param.Type.Fields))
				if param.ResolvedType != nil {
					fmt.Printf("   Debug: ResolvedType kind %v, fields %d\n", param.ResolvedType.Kind, len(param.ResolvedType.Fields))
				} else {
					fmt.Printf("   Debug: ResolvedType is nil\n")
				}
				
				// Check registry
				registry := parser.GetRegistry()
				if _, exists := registry.GetType("Company"); exists {
					fmt.Printf("   Debug: Company exists in registry\n")
				} else {
					fmt.Printf("   Debug: Company NOT in registry\n")
				}
			}
		}

		// Create API description
		desc := apigen.NewDescription("DemoAPI", enriched)

		// Generate JSON
		content, err := jsonGenerator.Generate(desc)
		if err != nil {
			log.Fatalf("Failed to generate JSON for %s: %v", f.name, err)
		}

		// Save to demo file using new API
		filename := filepath.Join(outputDir, fmt.Sprintf("%s.json", f.name))
		content.FileName = filename

		// Write using io.WriterTo
		file, err := os.Create(filename)
		if err != nil {
			log.Fatalf("Failed to create file %s: %v", filename, err)
		}
		defer file.Close()

		_, err = content.WriteTo(file)
		if err != nil {
			log.Fatalf("Failed to write demo file for %s: %v", f.name, err)
		}

		fmt.Printf("   Generated demo file: %s (%d methods)\n", filename, len(enriched))
	}

	// Generate individual method files for detailed inspection
	fmt.Println("\n3. Generating individual method demo files...")
	allEnriched, err := transformer.Transform(methods)
	if err != nil {
		log.Fatalf("Failed to transform all methods: %v", err)
	}

	desc := apigen.NewDescription("GoldenFileDemo", allEnriched)

	for methodName, methodDesc := range desc.Methods {
		// Create individual API description for this method
		singleMethodDesc := apigen.APIDescription{
			APIName:  "DemoAPI",
			Methods:  map[string]apigen.MethodDescription{methodName: methodDesc},
		}

		content, err := jsonGenerator.Generate(singleMethodDesc)
		if err != nil {
			log.Fatalf("Failed to generate JSON for method %s: %v", methodName, err)
		}

		filename := filepath.Join(outputDir, fmt.Sprintf("method_%s.json", methodName))
		content.FileName = filename

		// Write using io.WriterTo
		file, err := os.Create(filename)
		if err != nil {
			log.Fatalf("Failed to create file %s: %v", filename, err)
		}
		defer file.Close()

		_, err = content.WriteTo(file)
		if err != nil {
			log.Fatalf("Failed to write golden file for method %s: %v", methodName, err)
		}

		fmt.Printf("   Generated method file: %s\n", filename)
	}

	fmt.Println("\n4. Demo file generation complete!")
	fmt.Printf("   Check the '%s' directory to visually inspect the generated API descriptions.\n", outputDir)
	fmt.Println("   Each file shows how different Go types are represented in the JSON API format.")

	// List the generated files
	files, err := filepath.Glob(filepath.Join(outputDir, "*.json"))
	if err == nil {
		fmt.Println("\nGenerated files:")
		for _, file := range files {
			fmt.Printf("   - %s\n", file)
		}
	}
}