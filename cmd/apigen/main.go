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
		outputFile  = flag.String("out", "", "Output Go file path (cannot use with -stdout)")
		stdout      = flag.Bool("stdout", false, "Write to stdout instead of file")
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
	if !*stdout && *outputFile == "" {
		log.Fatal("either output file (-out) or stdout (-stdout) is required")
	}
	if *stdout && *outputFile != "" {
		log.Fatal("cannot specify both -out and -stdout, choose one")
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

	// Build configuration
	config := apigen.NewConfig().
		WithConstName(*constName)

	if *apiName != "" {
		config = config.WithAPIName(*apiName)
	}

	// Set input source
	if *packagePath != "" {
		config = config.WithPackage(*packagePath)
		// Infer API name from package if not specified
		if *apiName == "" {
			config = config.WithAPIName(inferPackageName(*packagePath))
		}
	} else {
		config = config.WithFile(*filePath)
		// Infer package name from file
		if inferredPkg := inferPackageFromFile(*filePath); inferredPkg != "" {
			config = config.WithPackageName(inferredPkg)
		}
	}

	// Set output target
	if *stdout {
		config = config.WithOutput(apigen.Stdout())
	} else {
		config = config.WithOutput(apigen.File(*outputFile))
	}

	// Determine package name for generators
	packageName := "main" // default
	if *packagePath != "" {
		// For package mode, infer from path
		if inferred := inferPackageName(*packagePath); inferred != "" {
			packageName = inferred
		}
	} else if *filePath != "" {
		// For file mode, read from file
		if inferred := inferPackageFromFile(*filePath); inferred != "" {
			packageName = inferred
		}
	}

	// Set generator type
	if *mapOutput {
		config = config.WithGenerator(apigen.NewGoMapGenerator(packageName, *constName))
	} else {
		config = config.WithGenerator(apigen.NewGoConstGenerator(packageName, *constName))
	}

	// Add filters
	if *methodList != "" {
		names := parseCommaSeparated(*methodList)
		config = config.WithMethodFilter(apigen.FilterByListFunc(names))
	}
	if *prefix != "" {
		config = config.WithMethodFilter(apigen.FilterByPrefixFunc(*prefix))
	}
	if *suffix != "" {
		config = config.WithMethodFilter(apigen.FilterBySuffixFunc(*suffix))
	}
	if *contains != "" {
		config = config.WithMethodFilter(apigen.FilterByContainsFunc(*contains))
	}

	// Generate
	err := apigen.Generate(config)
	if err != nil {
		log.Fatalf("Failed to generate: %v", err)
	}

	if !*stdout {
		fmt.Printf("Generated %s with constant %s\n", *outputFile, *constName)
	}
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
	if packagePath == "" {
		return "main"
	}
	parts := strings.Split(packagePath, "/")
	if len(parts) > 0 && parts[len(parts)-1] != "" {
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
	fmt.Println("  -out string        Output Go file path (cannot use with -stdout)")
	fmt.Println("  -stdout            Write to stdout instead of file")
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
	fmt.Println("  # Generate from package to stdout")
	fmt.Println("  go run github.com/pangobit/agent-sdk/cmd/apigen -package=./pkg/handlers -prefix=Handle -stdout -const=APIJSON")
	fmt.Println()
	fmt.Println("  # Generate from file with method list")
	fmt.Println("  go run github.com/pangobit/agent-sdk/cmd/apigen -file=handlers.go -methods=Method1,Method2 -out=api_gen.go -const=APIJSON")
	fmt.Println()
	fmt.Println("  # Use with go:generate")
	fmt.Println("  //go:generate go run github.com/pangobit/agent-sdk/cmd/apigen -file=main.go -prefix=Handle -out=api_gen.go -const=APIJSON")
}
