package main

import (
	"fmt"
	"github.com/pangobit/agent-sdk/pkg/apigen"
)

func debug() {
	parser := apigen.NewParser()
	methods, err := parser.ParsePackage("./api")
	if err != nil {
		panic(err)
	}

	registry := parser.GetRegistry()
	fmt.Printf("Found %d methods\n", len(methods))
	fmt.Printf("Registry has types:")
	
	// Try to get some known types
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
}