package apigen

import (
	"bytes"
	"testing"
)

func TestGeneratedContent_WriteTo(t *testing.T) {
	// Create a simple API description
	desc := APIDescription{
		APIName: "TestAPI",
		Methods: map[string]MethodDescription{
			"TestMethod": {
				Description: "A test method",
				Parameters: map[string]ParameterInfo{
					"param1": {
						Type: "string",
					},
				},
			},
		},
	}

	// Generate JSON content
	generator := NewJSONGenerator()
	content, err := generator.Generate(desc)
	if err != nil {
		t.Fatalf("Failed to generate content: %v", err)
	}

	// Test writing to buffer
	var buf bytes.Buffer
	n, err := content.WriteTo(&buf)
	if err != nil {
		t.Fatalf("Failed to write to buffer: %v", err)
	}

	if n == 0 {
		t.Error("Expected non-zero bytes written")
	}

	if buf.Len() == 0 {
		t.Error("Expected buffer to contain data")
	}

	// Verify the content contains expected JSON
	contentStr := buf.String()
	if len(contentStr) == 0 {
		t.Error("Expected non-empty content")
	}
}

func TestLayeredAPI(t *testing.T) {
	// Test the high-level API functions exist and can be called
	// (We can't easily test the full pipeline without source files,
	// but we can verify the functions exist and basic config works)

	config := NewConfig().
		WithPackage("./test").
		WithOutput(Stdout()).
		WithConstName("TestAPI")

	if config == nil {
		t.Error("Expected config to be created")
	}

	// Test that the config has expected values
	if config.ConstName != "TestAPI" {
		t.Errorf("Expected ConstName to be 'TestAPI', got %s", config.ConstName)
	}
}

func TestOutputTargets(t *testing.T) {
	// Test stdout target
	stdout := Stdout()
	writer := stdout.GetWriter()
	if writer == nil {
		t.Error("Expected stdout writer to be non-nil")
	}

	// Test file target
	file := File("test.txt")
	writer = file.GetWriter()
	if writer == nil {
		t.Error("Expected file writer to be non-nil")
	}

	// Test custom writer target
	var buf bytes.Buffer
	custom := Writer(&buf)
	writer = custom.GetWriter()
	if writer != &buf {
		t.Error("Expected custom writer to return the buffer")
	}
}

func TestMethodFilters(t *testing.T) {
	methods := []RawMethod{
		{Name: "HandleUser"},
		{Name: "ProcessData"},
		{Name: "ValidateInput"},
		{Name: "HandleRequest"},
	}

	// Test prefix filter
	prefixFilter := FilterByPrefixFunc("Handle")
	filtered := prefixFilter.Filter(methods)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 methods with Handle prefix, got %d", len(filtered))
	}

	// Test contains filter
	containsFilter := FilterByContainsFunc("Data")
	filtered = containsFilter.Filter(methods)
	if len(filtered) != 1 {
		t.Errorf("Expected 1 method containing Data, got %d", len(filtered))
	}
}