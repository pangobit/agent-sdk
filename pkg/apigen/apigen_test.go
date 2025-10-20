package apigen

import (
	"os"
	"strings"
	"testing"
)

// TestNewAPI tests the new v2.0 API
func TestNewAPI(t *testing.T) {
	// Create a temporary test file
	testFile := "test_new_api.go"
	testContent := `package test

// HandleEcho handles echo requests
func HandleEcho(message string, count int) error {
	return nil
}

// HandlePing handles ping requests
func HandlePing() {
}
`
	err := writeTestFile(testFile, testContent)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer removeTestFile(testFile)

	// Test the new API
	parser := NewParser()
	transformer := NewTransformer(NewTypeRegistry())
	generator := NewJSONGenerator()

	// Parse
	methods, err := parser.ParseSingleFile(testFile)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	// Filter
	filtered := FilterByPrefix(methods, "Handle")
	if len(filtered) != 2 {
		t.Errorf("expected 2 methods, got %d", len(filtered))
	}

	// Transform
	enriched, err := transformer.Transform(filtered)
	if err != nil {
		t.Fatalf("failed to transform methods: %v", err)
	}

	// Generate description
	desc := NewDescription("TestAPI", enriched)

	// Generate output
	content, err := generator.Generate(desc)
	if err != nil {
		t.Fatalf("failed to generate content: %v", err)
	}

	// Verify output contains expected data
	if !strings.Contains(content.Content, "TestAPI") {
		t.Error("generated content does not contain API name")
	}
	if !strings.Contains(content.Content, "HandleEcho") {
		t.Error("generated content does not contain HandleEcho method")
	}
	if !strings.Contains(content.Content, "HandlePing") {
		t.Error("generated content does not contain HandlePing method")
	}
}

// Helper functions for testing
func writeTestFile(filename, content string) error {
	return writeFile(filename, content)
}

func removeTestFile(filename string) {
	os.Remove(filename)
}

// writeFile is a helper to write test files
func writeFile(filename, content string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	return err
}