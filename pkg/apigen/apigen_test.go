package apigen

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestGenerateFromFile(t *testing.T) {
	tests := []struct {
		name        string
		config      GeneratorConfig
		expectError bool
		expectCount int
	}{
		{
			name: "prefix_filter",
			config: WithPrefix("Handle").SetAPIName("TestAPI"),
			expectError: false,
			expectCount: 2, // HandleEcho, HandlePing
		},
		{
			name: "suffix_filter",
			config: WithSuffix("Handler").SetAPIName("TestAPI"),
			expectError: false,
			expectCount: 1, // SomeHandler
		},
		{
			name: "contains_filter",
			config: WithContains("Echo").SetAPIName("TestAPI"),
			expectError: false,
			expectCount: 1, // HandleEcho
		},
		{
			name: "method_list_filter",
			config: WithMethodList("HandleEcho", "SomeHandler").SetAPIName("TestAPI"),
			expectError: false,
			expectCount: 2,
		},
		{
			name: "exclude_http_types",
			config: WithPrefix("HTTP").SetAPIName("TestAPI"),
			expectError: false,
			expectCount: 1, // HTTPMethod should exclude http.Request parameter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary test file
			testFile := "test_sample.go"
			testContent := `package test

import "net/http"

// HandleEcho handles echo requests
// Parameters:
//   - message: the message to echo
func HandleEcho(message string, count int) error {
	return nil
}

// HandlePing handles ping requests
func HandlePing() {
}

// SomeHandler is a general handler
func SomeHandler(data []byte, config map[string]interface{}) error {
	return nil
}

// HTTPMethod demonstrates HTTP parameter exclusion
func HTTPMethod(req *http.Request, data string) error {
	return nil
}

// PrivateMethod should not be included in prefix filter
func privateMethod() {
}
`
			err := writeTestFile(testFile, testContent)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}
			defer removeTestFile(testFile)

			result, err := GenerateFromFile(testFile, tt.config)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if err != nil {
				return
			}

			if result.APIName != tt.config.APIName {
				t.Errorf("expected API name %s, got %s", tt.config.APIName, result.APIName)
			}

			if len(result.Methods) != tt.expectCount {
				t.Errorf("expected %d methods, got %d", tt.expectCount, len(result.Methods))
			}

			// Verify JSON serialization works
			jsonData, err := json.Marshal(result)
			if err != nil {
				t.Errorf("failed to marshal to JSON: %v", err)
			}
			if len(jsonData) == 0 {
				t.Error("JSON output is empty")
			}
		})
	}
}

func TestMethodDescriptionGeneration(t *testing.T) {
	testFile := "test_method.go"
	testContent := `package test

// ProcessData processes the given data
// This function demonstrates parameter extraction
func ProcessData(name string, age int, items []string, config map[string]interface{}) error {
	return nil
}
`
	err := writeTestFile(testFile, testContent)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer removeTestFile(testFile)

	config := WithMethodList("ProcessData").SetAPIName("TestAPI")
	result, err := GenerateFromFile(testFile, config)
	if err != nil {
		t.Fatalf("failed to generate description: %v", err)
	}

	method, exists := result.Methods["ProcessData"]
	if !exists {
		t.Fatal("ProcessData method not found")
	}

	if method.Description != "ProcessData processes the given data This function demonstrates parameter extraction" {
		t.Errorf("unexpected description: %s", method.Description)
	}

	expectedParams := map[string]string{
		"name":   "string",
		"age":    "int",
		"items":  "[]string",
		"config": "map[string]interface{}",
	}

	if len(method.Parameters) != len(expectedParams) {
		t.Errorf("expected %d parameters, got %d", len(expectedParams), len(method.Parameters))
	}

	for paramName, expectedType := range expectedParams {
		param, exists := method.Parameters[paramName]
		if !exists {
			t.Errorf("parameter %s not found", paramName)
			continue
		}
		if param.Type != expectedType {
			t.Errorf("parameter %s: expected type %s, got %s", paramName, expectedType, param.Type)
		}
	}
}

func TestHTTPParameterExclusion(t *testing.T) {
	testFile := "test_http.go"
	testContent := `package test

import "net/http"

// HandleRequest handles HTTP requests
func HandleRequest(w http.ResponseWriter, req *http.Request, userID string) error {
	return nil
}
`
	err := writeTestFile(testFile, testContent)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer removeTestFile(testFile)

	config := WithMethodList("HandleRequest").SetAPIName("TestAPI")
	result, err := GenerateFromFile(testFile, config)
	if err != nil {
		t.Fatalf("failed to generate description: %v", err)
	}

	method := result.Methods["HandleRequest"]

	// Should only have userID parameter, not w or req
	if len(method.Parameters) != 1 {
		t.Errorf("expected 1 parameter after HTTP exclusion, got %d", len(method.Parameters))
	}

	param, exists := method.Parameters["userID"]
	if !exists {
		t.Error("userID parameter not found")
	} else if param.Type != "string" {
		t.Errorf("expected userID type string, got %s", param.Type)
	}
}

func TestFilterHelpers(t *testing.T) {
	prefixConfig := WithPrefix("Test")
	if prefixConfig.Strategy != StrategyPrefix || prefixConfig.Filter != "Test" {
		t.Error("WithPrefix not configured correctly")
	}

	suffixConfig := WithSuffix("Handler")
	if suffixConfig.Strategy != StrategySuffix || suffixConfig.Filter != "Handler" {
		t.Error("WithSuffix not configured correctly")
	}

	containsConfig := WithContains("Proc")
	if containsConfig.Strategy != StrategyContains || containsConfig.Filter != "Proc" {
		t.Error("WithContains not configured correctly")
	}

	methodListConfig := WithMethodList("Method1", "Method2")
	if len(methodListConfig.MethodList) != 2 {
		t.Error("WithMethodList not configured correctly")
	}
}

func TestGenerateAndWriteGoFile(t *testing.T) {
	// Create a temporary test file
	testFile := "test_gen.go"
	testContent := `package test

// ProcessData processes data
func ProcessData(input string) error {
	return nil
}
`
	err := writeTestFile(testFile, testContent)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer removeTestFile(testFile)

	outputFile := "test_api_gen.go"
	defer removeTestFile(outputFile)

	config := WithMethodList("ProcessData").SetAPIName("TestAPI")

	err = GenerateAndWriteGoFileFromFile(testFile, outputFile, "TestAPIJSON", "test", config)
	if err != nil {
		t.Fatalf("failed to generate Go file: %v", err)
	}

	// Check that the file was created and contains expected content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Check for expected patterns
	expectedPatterns := []string{
		"// Code generated by apigen; DO NOT EDIT.",
		"package test",
		"const TestAPIJSON = `",
		`"apiName": "TestAPI"`,
		`"ProcessData"`,
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(contentStr, pattern) {
			t.Errorf("generated file does not contain expected pattern: %s", pattern)
		}
	}

	// Check that it's valid JSON by extracting and parsing the constant
	start := strings.Index(contentStr, "const TestAPIJSON = `")
	if start == -1 {
		t.Fatal("could not find constant declaration")
	}
	start += len("const TestAPIJSON = `")

	end := strings.LastIndex(contentStr, "`")
	if end == -1 {
		t.Fatal("could not find end of constant")
	}

	jsonStr := contentStr[start:end]

	var desc APIDescription
	err = json.Unmarshal([]byte(jsonStr), &desc)
	if err != nil {
		t.Fatalf("generated JSON is invalid: %v", err)
	}

	if desc.APIName != "TestAPI" {
		t.Errorf("expected API name 'TestAPI', got '%s'", desc.APIName)
	}
}

func TestGenerateAndWriteGoFileAsMap(t *testing.T) {
	// Create a temporary test file
	testFile := "test_gen_map.go"
	testContent := `package test

// ProcessData processes data
func ProcessData(input string) error {
	return nil
}

// HandleRequest handles a request
func HandleRequest(id int, name string) error {
	return nil
}
`
	err := writeTestFile(testFile, testContent)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer removeTestFile(testFile)

	outputFile := "test_api_gen_map.go"
	defer removeTestFile(outputFile)

	config := WithMethodList("ProcessData", "HandleRequest").SetAPIName("TestAPI")

	err = GenerateAndWriteGoFileFromFileAsMap(testFile, outputFile, "APIDefinitions", "test", config)
	if err != nil {
		t.Fatalf("failed to generate Go file: %v", err)
	}

	// Check that the file was created and contains expected content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Check for expected patterns
	expectedPatterns := []string{
		"// Code generated by apigen; DO NOT EDIT.",
		"package test",
		"var APIDefinitions = map[string]string{",
		`"ProcessData": `,
		`"HandleRequest": `,
		`"description":`,
		`"parameters":`,
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(contentStr, pattern) {
			t.Errorf("generated file does not contain expected pattern: %s", pattern)
		}
	}

	// Check that it's valid Go code by attempting to parse it
	// (We can't actually compile it without proper imports, but we can check syntax)
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, outputFile, content, parser.ParseComments)
	if err != nil {
		t.Errorf("generated file is not valid Go syntax: %v", err)
	}

	// Extract and verify the map structure
	start := strings.Index(contentStr, "var APIDefinitions = map[string]string{")
	if start == -1 {
		t.Fatal("could not find map variable declaration")
	}

	end := strings.LastIndex(contentStr, "}")
	if end == -1 {
		t.Fatal("could not find end of map")
	}

	mapContent := contentStr[start:end+1]

	// Should contain both method keys
	if !strings.Contains(mapContent, `"ProcessData":`) {
		t.Error("map does not contain ProcessData key")
	}
	if !strings.Contains(mapContent, `"HandleRequest":`) {
		t.Error("map does not contain HandleRequest key")
	}

	// Should contain JSON content (look for description and parameters)
	if !strings.Contains(mapContent, `"description"`) {
		t.Error("map values do not contain description field")
	}
	if !strings.Contains(mapContent, `"parameters"`) {
		t.Error("map values do not contain parameters field")
	}

	// Should NOT contain apiName (since we're omitting it)
	if strings.Contains(mapContent, `"apiName"`) {
		t.Error("map should not contain apiName field")
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