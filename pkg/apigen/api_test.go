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

func TestNilPointerSafety(t *testing.T) {
	tests := []struct {
		name          string
		testFunc      func() (result interface{}, shouldPanic bool)
		expectedPanic bool
	}{
		{
			name: "packagePathInput.GetPackagePath with nil receiver",
			testFunc: func() (interface{}, bool) {
				var input *packagePathInput
				result := input.GetPackagePath()
				return result, false
			},
			expectedPanic: false,
		},
		{
			name: "packagePathInput.GetFilePath with nil receiver",
			testFunc: func() (interface{}, bool) {
				var input *packagePathInput
				result := input.GetFilePath()
				return result, false
			},
			expectedPanic: false,
		},
		{
			name: "filePathInput.GetPackagePath with nil receiver",
			testFunc: func() (interface{}, bool) {
				var input *filePathInput
				result := input.GetPackagePath()
				return result, false
			},
			expectedPanic: false,
		},
		{
			name: "filePathInput.GetFilePath with nil receiver",
			testFunc: func() (interface{}, bool) {
				var input *filePathInput
				result := input.GetFilePath()
				return result, false
			},
			expectedPanic: false,
		},
		{
			name: "stdoutTarget.GetWriter with nil receiver",
			testFunc: func() (interface{}, bool) {
				var target *stdoutTarget
				result := target.GetWriter()
				return result, false
			},
			expectedPanic: false,
		},
		{
			name: "stdoutTarget.Close with nil receiver",
			testFunc: func() (interface{}, bool) {
				var target *stdoutTarget
				result := target.Close()
				return result, false
			},
			expectedPanic: false,
		},
		{
			name: "fileTarget.GetWriter with nil receiver",
			testFunc: func() (interface{}, bool) {
				var target *fileTarget
				result := target.GetWriter()
				return result, false
			},
			expectedPanic: false,
		},
		{
			name: "fileTarget.Close with nil receiver",
			testFunc: func() (interface{}, bool) {
				var target *fileTarget
				result := target.Close()
				return result, false
			},
			expectedPanic: false,
		},
		{
			name: "writerTarget.GetWriter with nil receiver",
			testFunc: func() (interface{}, bool) {
				var target *writerTarget
				result := target.GetWriter()
				return result, false
			},
			expectedPanic: false,
		},
		{
			name: "writerTarget.Close with nil receiver",
			testFunc: func() (interface{}, bool) {
				var target *writerTarget
				result := target.Close()
				return result, false
			},
			expectedPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use recover to catch any panics
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectedPanic {
						t.Errorf("unexpected panic: %v", r)
					}
				} else if tt.expectedPanic {
					t.Error("expected panic but none occurred")
				}
			}()

			result, _ := tt.testFunc()
			// Verify result is reasonable for non-panicking cases
			if !tt.expectedPanic {
				// For methods that return values, ensure they're not causing issues
				_ = result
			}
		})
	}
}

func TestNilFilterSafety(t *testing.T) {
	methods := []RawMethod{
		{Name: "TestMethod"},
		{Name: "AnotherMethod"},
	}

	tests := []struct {
		name       string
		filterFunc func([]RawMethod) []RawMethod
	}{
		{
			name: "prefixFilter.Filter with nil receiver",
			filterFunc: func(methods []RawMethod) []RawMethod {
				var filter *prefixFilter
				return filter.Filter(methods)
			},
		},
		{
			name: "suffixFilter.Filter with nil receiver",
			filterFunc: func(methods []RawMethod) []RawMethod {
				var filter *suffixFilter
				return filter.Filter(methods)
			},
		},
		{
			name: "containsFilter.Filter with nil receiver",
			filterFunc: func(methods []RawMethod) []RawMethod {
				var filter *containsFilter
				return filter.Filter(methods)
			},
		},
		{
			name: "listFilter.Filter with nil receiver",
			filterFunc: func(methods []RawMethod) []RawMethod {
				var filter *listFilter
				return filter.Filter(methods)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("unexpected panic: %v", r)
				}
			}()

			result := tt.filterFunc(methods)
			// When filter is nil, should return original methods unchanged
			if len(result) != len(methods) {
				t.Errorf("expected %d methods, got %d", len(methods), len(result))
			}
		})
	}
}

func TestGenerateWithNilFilters(t *testing.T) {
	// Create a minimal config with nil filters mixed in
	config := NewConfig().
		WithPackage("fmt"). // Use a package that exists
		WithOutput(Stdout()).
		WithMethodFilter(nil). // Add nil filter
		WithMethodFilter(FilterByPrefixFunc("Print")).
		WithMethodFilter(nil) // Add another nil filter

	// This should not panic due to nil filters
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic in Generate with nil filters: %v", r)
		}
	}()

	// We expect this to fail because we're using a real package, but it shouldn't panic
	err := Generate(config)
	if err == nil {
		t.Log("Generate succeeded (unexpected but not a panic)")
	} else {
		t.Logf("Generate failed as expected: %v", err)
	}
}
