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
			name:        "prefix_filter",
			config:      WithPrefix("Handle").SetAPIName("TestAPI"),
			expectError: false,
			expectCount: 2, // HandleEcho, HandlePing
		},
		{
			name:        "suffix_filter",
			config:      WithSuffix("Handler").SetAPIName("TestAPI"),
			expectError: false,
			expectCount: 1, // SomeHandler
		},
		{
			name:        "contains_filter",
			config:      WithContains("Echo").SetAPIName("TestAPI"),
			expectError: false,
			expectCount: 1, // HandleEcho
		},
		{
			name:        "method_list_filter",
			config:      WithMethodList("HandleEcho", "SomeHandler").SetAPIName("TestAPI"),
			expectError: false,
			expectCount: 2,
		},
		{
			name:        "exclude_http_types",
			config:      WithPrefix("HTTP").SetAPIName("TestAPI"),
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

	mapContent := contentStr[start : end+1]

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

func TestSliceOfStructParsing(t *testing.T) {
	testFile := "test_slice_struct.go"
	testContent := `package test

// Item represents an item in a collection
type Item struct {
	ID       int    ` + "`" + `json:"id"` + "`" + `
	Name     string ` + "`" + `json:"name"` + "`" + `
	Category string ` + "`" + `json:"category"` + "`" + `
}

// Collection represents a collection with nested slice of items
type Collection struct {
	Title string ` + "`" + `json:"title"` + "`" + `
	Items []Item ` + "`" + `json:"items"` + "`" + `
	Data  string ` + "`" + `json:"data"` + "`" + `
}

// ProcessCollection processes collection data
func ProcessCollection(coll Collection) error {
	return nil
}
`
	err := writeTestFile(testFile, testContent)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer removeTestFile(testFile)

	config := WithMethodList("ProcessCollection").SetAPIName("TestAPI")
	result, err := GenerateFromFile(testFile, config)
	if err != nil {
		t.Fatalf("failed to generate description: %v", err)
	}

	method, exists := result.Methods["ProcessCollection"]
	if !exists {
		t.Fatal("ProcessCollection method not found")
	}

	collParam, exists := method.Parameters["coll"]
	if !exists {
		t.Fatal("coll parameter not found")
	}

	if collParam.Type != "Collection" {
		t.Errorf("expected coll type 'Collection', got '%s'", collParam.Type)
	}

	// Check that Collection fields exist
	collectionFields := collParam.Fields
	if collectionFields == nil {
		t.Fatal("Collection should have fields")
	}

	// Check Collection.Items field (slice of structs)
	itemsField, exists := collectionFields["Items"]
	if !exists {
		t.Error("Collection.Items field not found")
	} else {
		if itemsField.Type != "[]Item" {
			t.Errorf("Collection.Items: expected type '[]Item', got '%s'", itemsField.Type)
		}
		if itemsField.Annotations["json"] != "items" {
			t.Errorf("Collection.Items: expected json annotation 'items', got '%s'", itemsField.Annotations["json"])
		}

		// Check that the slice element type (Item) fields are introspected
		itemFields := itemsField.Fields
		if itemFields == nil {
			t.Fatal("Items slice should have introspected Item fields")
		}

		expectedItemFields := map[string]string{
			"ID":       "int",
			"Name":     "string",
			"Category": "string",
		}

		for fieldName, expectedType := range expectedItemFields {
			field, exists := itemFields[fieldName]
			if !exists {
				t.Errorf("Item.%s field not found", fieldName)
			} else {
				if field.Type != expectedType {
					t.Errorf("Item.%s: expected type '%s', got '%s'", fieldName, expectedType, field.Type)
				}
			}
		}
	}
}

func TestCrossPackageNestedStructParsing(t *testing.T) {
	// Create a test file with types from different packages
	testFile := "test_cross_package.go"
	testContent := `package test

import "time"

type ExternalPackageType struct {
	ID       int                    ` + "`json:\"id\"`" + `
	Name     string                 ` + "`json:\"name\"`" + `
	Metadata map[string]interface{} ` + "`json:\"metadata\"`" + `
	Created  time.Time              ` + "`json:\"created\"`" + `
}

type MyStruct struct {
	Test ExternalPackageType ` + "`json:\"test\"`" + `
	Data string              ` + "`json:\"data\"`" + `
}

func ProcessMyStruct(ms MyStruct) error {
	return nil
}
`
	err := writeTestFile(testFile, testContent)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer removeTestFile(testFile)

	config := WithMethodList("ProcessMyStruct").SetAPIName("TestAPI")
	result, err := GenerateFromFile(testFile, config)
	if err != nil {
		t.Fatalf("failed to generate description: %v", err)
	}

	method, exists := result.Methods["ProcessMyStruct"]
	if !exists {
		t.Fatal("ProcessMyStruct method not found")
	}

	msParam, exists := method.Parameters["ms"]
	if !exists {
		t.Fatal("ms parameter not found")
	}

	if msParam.Type != "MyStruct" {
		t.Errorf("expected ms type 'MyStruct', got '%s'", msParam.Type)
	}

	// Check that MyStruct fields exist
	myStructFields := msParam.Fields
	if myStructFields == nil {
		t.Fatal("MyStruct should have fields")
	}

	// Check MyStruct.Data field
	dataField, exists := myStructFields["Data"]
	if !exists {
		t.Error("MyStruct.Data field not found")
	} else {
		if dataField.Type != "string" {
			t.Errorf("MyStruct.Data: expected type 'string', got '%s'", dataField.Type)
		}
		if dataField.Annotations["json"] != "data" {
			t.Errorf("MyStruct.Data: expected json annotation 'data', got '%s'", dataField.Annotations["json"])
		}
	}

	// Check MyStruct.Test field (nested struct from same package)
	testField, exists := myStructFields["Test"]
	if !exists {
		t.Error("MyStruct.Test field not found")
	} else {
		if testField.Type != "ExternalPackageType" {
			t.Errorf("MyStruct.Test: expected type 'ExternalPackageType', got '%s'", testField.Type)
		}
		if testField.Annotations["json"] != "test" {
			t.Errorf("MyStruct.Test: expected json annotation 'test', got '%s'", testField.Annotations["json"])
		}

		// Check nested ExternalPackageType fields
		externalFields := testField.Fields
		if externalFields == nil {
			t.Fatal("ExternalPackageType should have nested fields")
		}

		expectedFields := map[string]string{
			"ID":       "int",
			"Name":     "string",
			"Metadata": "map[string]interface{}",
			"Created":  "time.Time",
		}

		for fieldName, expectedType := range expectedFields {
			field, exists := externalFields[fieldName]
			if !exists {
				t.Errorf("ExternalPackageType.%s field not found", fieldName)
			} else {
				if field.Type != expectedType {
					t.Errorf("ExternalPackageType.%s: expected type '%s', got '%s'", fieldName, expectedType, field.Type)
				}
			}
		}
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
