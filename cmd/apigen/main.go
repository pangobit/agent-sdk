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
		packagePath = flag.String("package", "", "Package path to analyze (for package mode)")
		filePath    = flag.String("file", "", "Go file to analyze (for file mode)")
		outputFile  = flag.String("out", "", "Output Go file path")
		constName   = flag.String("const", "", "Name of the generated constant")
		apiName     = flag.String("api-name", "", "Name for the generated API")
		methodList  = flag.String("methods", "", "Comma-separated list of method names to include")
		prefix      = flag.String("prefix", "", "Include methods starting with this prefix")
		suffix      = flag.String("suffix", "", "Include methods ending with this suffix")
		contains    = flag.String("contains", "", "Include methods containing this string")
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

	// Build configuration
	config := buildConfig(*apiName, *methodList, *prefix, *suffix, *contains)

	// Generate the file
	var err error
	if *packagePath != "" {
		err = apigen.GenerateAndWriteGoFile(*packagePath, *outputFile, *constName, packageName, config)
	} else {
		err = apigen.GenerateAndWriteGoFileFromFile(*filePath, *outputFile, *constName, packageName, config)
	}

	if err != nil {
		log.Fatalf("Failed to generate API file: %v", err)
	}

	fmt.Printf("Generated %s with constant %s\n", *outputFile, *constName)
}

func buildConfig(apiName, methodList, prefix, suffix, contains string) apigen.GeneratorConfig {
	var config apigen.GeneratorConfig

	// Set API name if provided
	if apiName != "" {
		config.APIName = apiName
	}

	// Apply filtering strategy
	if methodList != "" {
		// Parse comma-separated method list
		methods := parseCommaSeparated(methodList)
		config = apigen.WithMethodList(methods...).SetAPIName(apiName)
	} else if prefix != "" {
		config = apigen.WithPrefix(prefix).SetAPIName(apiName)
	} else if suffix != "" {
		config = apigen.WithSuffix(suffix).SetAPIName(apiName)
	} else if contains != "" {
		config = apigen.WithContains(contains).SetAPIName(apiName)
	} else {
		// Default: include all methods (no filtering)
		config = apigen.GeneratorConfig{ExcludeHTTP: true}
		if apiName != "" {
			config.APIName = apiName
		}
	}

	return config
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
