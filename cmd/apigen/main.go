package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pangobit/agent-sdk/pkg/apigen"
)

func main() {
	var (
		packagePath = flag.String("package", "", "Package path to analyze (cannot use with -file)")
		filePath    = flag.String("file", "", "Go file to analyze (cannot use with -package)")
		outputFile  = flag.String("out", "", "Output Go file path")
		constName   = flag.String("const", "", "Name of the generated constant")
		apiName     = flag.String("api-name", "", "Name for the generated API")
		methodList  = flag.String("methods", "", "Comma-separated list of method names to include")
		prefix      = flag.String("prefix", "", "Include methods starting with this prefix")
		suffix      = flag.String("suffix", "", "Include methods ending with this suffix")
		contains    = flag.String("contains", "", "Include methods containing this string")
		mapOutput   = flag.Bool("map", false, "Generate map[string]string instead of single JSON string")
		help        = flag.Bool("help", false, "Show help")
	)

	flag.Parse()

	if *help {
		showHelp()
		os.Exit(0)
	}

	// Validate required parameters
	if *outputFile == "" {
		log.Fatal("output file (-out) is required")
	}
	if *constName == "" {
		log.Fatal("constant name (-const) is required")
	}
	if *packagePath == "" && *filePath == "" {
		log.Fatal("either package path (-package) or file path (-file) is required")
	}
	if *packagePath != "" && *filePath != "" {
		log.Fatal("cannot specify both -package and -file, choose one")
	}

	// Determine package name from file path if not specified
	packageName := "main" // default
	if *packagePath != "" {
		// For package mode, try to infer package name or use default
		if *apiName == "" {
			*apiName = inferPackageName(*packagePath)
		}
	} else if *filePath != "" {
		// For file mode, try to infer package name from file
		if inferredPkg := inferPackageFromFile(*filePath); inferredPkg != "" {
			packageName = inferredPkg
		}
	}

	// New v2.0 API pipeline
	parser := apigen.NewParser()
	transformer := apigen.NewTransformer(apigen.NewTypeRegistry())

	// Parse methods
	var methods []apigen.RawMethod
	var err error
	if *packagePath != "" {
		methods, err = parser.ParsePackage(*packagePath)
	} else {
		methods, err = parser.ParseSingleFile(*filePath)
	}
	if err != nil {
		log.Fatalf("Failed to parse: %v", err)
	}

	// Apply filtering
	filteredMethods := applyFiltering(methods, *methodList, *prefix, *suffix, *contains)

	// Transform methods
	enrichedMethods, err := transformer.Transform(filteredMethods)
	if err != nil {
		log.Fatalf("Failed to transform methods: %v", err)
	}

	// Create API description
	desc := apigen.NewDescription(*apiName, enrichedMethods)

	// Generate output
	var content apigen.GeneratedContent
	if *mapOutput {
		generator := apigen.NewGoMapGenerator(packageName, *constName)
		content, err = generator.Generate(desc)
	} else {
		generator := apigen.NewGoConstGenerator(packageName, *constName)
		content, err = generator.Generate(desc)
	}
	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}

	// Write to file
	writer := apigen.NewWriter()
	err = writer.WriteToFile(content, *outputFile)
	if err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	fmt.Printf("Generated %s with constant %s\n", *outputFile, *constName)
}

func applyFiltering(methods []apigen.RawMethod, methodList, prefix, suffix, contains string) []apigen.RawMethod {
	if methodList != "" {
		// Parse comma-separated method list
		names := parseCommaSeparated(methodList)
		return apigen.FilterByList(methods, names)
	} else if prefix != "" {
		return apigen.FilterByPrefix(methods, prefix)
	} else if suffix != "" {
		return apigen.FilterBySuffix(methods, suffix)
	} else if contains != "" {
		return apigen.FilterByContains(methods, contains)
	}

	// No filtering - return all methods
	return methods
}

func parseCommaSeparated(input string) []string {
	if input == "" {
		return nil
	}

	var result []string
	current := ""
	inQuotes := false

	for _, r := range input {
		switch r {
		case '"':
			inQuotes = !inQuotes
		case ',':
			if !inQuotes {
				if current != "" {
					result = append(result, current)
					current = ""
				}
			} else {
				current += string(r)
			}
		default:
			current += string(r)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

func inferPackageName(packagePath string) string {
	// Simple heuristic: use the last component of the path
	parts := strings.Split(packagePath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "main"
}

func inferPackageFromFile(filePath string) string {
	// Try to read the file and extract package declaration
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}

	lines := strings.SplitSeq(string(content), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}

	return ""
}

func showHelp() {
	fmt.Println("apigen - Generate Go files with embedded API descriptions")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  go run github.com/pangobit/agent-sdk/cmd/apigen [flags]")
	fmt.Println()
	fmt.Println("FLAGS:")
	fmt.Println("  -package string    Package path to analyze (cannot use with -file)")
	fmt.Println("  -file string       Go file to analyze (cannot use with -package)")
	fmt.Println("  -out string        Output Go file path (required)")
	fmt.Println("  -const string      Name of the generated constant (required)")
	fmt.Println("  -api-name string   Name for the generated API")
	fmt.Println("  -methods string    Comma-separated list of method names to include")
	fmt.Println("  -prefix string     Include methods starting with this prefix")
	fmt.Println("  -suffix string     Include methods ending with this suffix")
	fmt.Println("  -contains string   Include methods containing this string")
	fmt.Println("  -map               Generate map[string]string instead of single JSON string")
	fmt.Println("  -help              Show this help")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Generate from package with prefix filter")
	fmt.Println("  go run github.com/pangobit/agent-sdk/cmd/apigen -package=./pkg/handlers -prefix=Handle -out=api_gen.go -const=APIJSON")
	fmt.Println()
	fmt.Println("  # Generate from file with method list")
	fmt.Println("  go run github.com/pangobit/agent-sdk/cmd/apigen -file=handlers.go -methods=Method1,Method2 -out=api_gen.go -const=APIJSON")
	fmt.Println()
	fmt.Println("  # Use with go:generate")
	fmt.Println("  //go:generate go run github.com/pangobit/agent-sdk/cmd/apigen -file=main.go -prefix=Handle -out=api_gen.go -const=APIJSON")
}
