package api

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestPromptTemplateArgument_String(t *testing.T) {
	tests := []struct {
		name     string
		arg      *PromptTemplateArgument
		wantJSON bool
		wantType string
	}{
		{
			name: "string argument",
			arg: &PromptTemplateArgument{
				Name:     "message",
				Argument: "hello world",
				Required: true,
			},
			wantJSON: true,
			wantType: "string",
		},
		{
			name: "int argument",
			arg: &PromptTemplateArgument{
				Name:     "count",
				Argument: 42,
				Required: false,
			},
			wantJSON: true,
			wantType: "int",
		},
		{
			name: "float argument",
			arg: &PromptTemplateArgument{
				Name:     "price",
				Argument: 3.14159,
				Required: true,
			},
			wantJSON: true,
			wantType: "float64",
		},
		{
			name: "bool argument",
			arg: &PromptTemplateArgument{
				Name:     "enabled",
				Argument: true,
				Required: false,
			},
			wantJSON: true,
			wantType: "bool",
		},
		{
			name: "slice argument",
			arg: &PromptTemplateArgument{
				Name:     "items",
				Argument: []string{"item1", "item2", "item3"},
				Required: true,
			},
			wantJSON: true,
			wantType: "[]string",
		},
		{
			name: "map argument",
			arg: &PromptTemplateArgument{
				Name:     "config",
				Argument: map[string]interface{}{"key": "value", "number": 42},
				Required: false,
			},
			wantJSON: true,
			wantType: "map[string]interface {}",
		},
		{
			name: "struct argument",
			arg: &PromptTemplateArgument{
				Name: "user",
				Argument: struct {
					Name string
					Age  int
				}{Name: "John", Age: 30},
				Required: true,
			},
			wantJSON: true,
			wantType: "struct { Name string; Age int }",
		},
		{
			name: "nil argument",
			arg: &PromptTemplateArgument{
				Name:     "optional",
				Argument: nil,
				Required: false,
			},
			wantJSON: true,
			wantType: "null",
		},
		{
			name: "empty string argument",
			arg: &PromptTemplateArgument{
				Name:     "empty",
				Argument: "",
				Required: true,
			},
			wantJSON: true,
			wantType: "string",
		},
		{
			name: "zero value int argument",
			arg: &PromptTemplateArgument{
				Name:     "zero",
				Argument: 0,
				Required: false,
			},
			wantJSON: true,
			wantType: "int",
		},
		{
			name: "empty slice argument",
			arg: &PromptTemplateArgument{
				Name:     "emptySlice",
				Argument: []int{},
				Required: true,
			},
			wantJSON: true,
			wantType: "[]int",
		},
		{
			name: "empty map argument",
			arg: &PromptTemplateArgument{
				Name:     "emptyMap",
				Argument: map[string]string{},
				Required: false,
			},
			wantJSON: true,
			wantType: "map[string]string",
		},
		{
			name: "complex slice argument",
			arg: &PromptTemplateArgument{
				Name:     "complexSlice",
				Argument: []interface{}{"string", 42, true, nil},
				Required: true,
			},
			wantJSON: true,
			wantType: "[]interface {}",
		},
		{
			name: "complex map argument",
			arg: &PromptTemplateArgument{
				Name: "complexMap",
				Argument: map[string]interface{}{
					"string": "value",
					"number": 42,
					"bool":   true,
					"slice":  []int{1, 2, 3},
					"nested": map[string]string{"key": "value"},
				},
				Required: false,
			},
			wantJSON: true,
			wantType: "map[string]interface {}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.arg.String()

			// Test that result is valid JSON
			if tt.wantJSON {
				var parsed map[string]interface{}
				err := json.Unmarshal([]byte(result), &parsed)
				if err != nil {
					t.Errorf("String() returned invalid JSON: %v", err)
					return
				}

				// Test required fields are present
				if name, ok := parsed["name"].(string); !ok || name != tt.arg.Name {
					t.Errorf("String() name field = %v, want %v", name, tt.arg.Name)
				}

				if required, ok := parsed["required"].(bool); !ok || required != tt.arg.Required {
					t.Errorf("String() required field = %v, want %v", required, tt.arg.Required)
				}

				// Handle type field - can be string or null
				if tt.wantType == "null" {
					if parsed["type"] != nil {
						t.Errorf("String() type field = %v, want null", parsed["type"])
					}
				} else {
					if typeStr, ok := parsed["type"].(string); !ok || typeStr != tt.wantType {
						t.Errorf("String() type field = %v, want %v", typeStr, tt.wantType)
					}
				}

				// Test items field for slices and maps
				if tt.arg.Argument != nil && (reflect.TypeOf(tt.arg.Argument).Kind() == reflect.Slice ||
					reflect.TypeOf(tt.arg.Argument).Kind() == reflect.Map) {
					if items, ok := parsed["items"].(map[string]interface{}); !ok {
						t.Errorf("String() missing items field for slice/map type")
					} else {
						if itemType, ok := items["type"].(string); !ok || itemType != tt.wantType {
							t.Errorf("String() items.type field = %v, want %v", itemType, tt.wantType)
						}
					}
				} else {
					// For non-slice/map types, items should not be present or should be empty
					if items, exists := parsed["items"]; exists && items != nil {
						t.Errorf("String() should not have items field for non-slice/map type, got %v", items)
					}
				}
			} else {
				if result != "" {
					t.Errorf("String() should return empty string for invalid cases, got %v", result)
				}
			}
		})
	}
}

// TestPromptTemplateArgument_StringEdgeCases tests edge cases and boundary conditions
func TestPromptTemplateArgument_StringEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		arg  *PromptTemplateArgument
	}{
		{
			name: "very long string argument",
			arg: &PromptTemplateArgument{
				Name:     "longString",
				Argument: "This is a very long string that contains many characters and should be handled properly by the String method. It includes various types of content and should not cause any issues with the JSON marshaling.",
				Required: true,
			},
		},
		{
			name: "unicode string argument",
			arg: &PromptTemplateArgument{
				Name:     "unicodeString",
				Argument: "Hello 世界 🌍 🚀",
				Required: false,
			},
		},
		{
			name: "special characters in string",
			arg: &PromptTemplateArgument{
				Name:     "specialChars",
				Argument: "Special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?",
				Required: true,
			},
		},
		{
			name: "very large number",
			arg: &PromptTemplateArgument{
				Name:     "largeNumber",
				Argument: 999999999999999999,
				Required: false,
			},
		},
		{
			name: "negative number",
			arg: &PromptTemplateArgument{
				Name:     "negativeNumber",
				Argument: -42,
				Required: true,
			},
		},
		{
			name: "float with many decimal places",
			arg: &PromptTemplateArgument{
				Name:     "preciseFloat",
				Argument: 3.141592653589793238462643383279,
				Required: false,
			},
		},
		{
			name: "large slice",
			arg: &PromptTemplateArgument{
				Name:     "largeSlice",
				Argument: make([]int, 1000),
				Required: true,
			},
		},
		{
			name: "large map",
			arg: &PromptTemplateArgument{
				Name: "largeMap",
				Argument: func() map[string]int {
					m := make(map[string]int)
					for i := 0; i < 100; i++ {
						m[fmt.Sprintf("key%d", i)] = i
					}
					return m
				}(),
				Required: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.arg.String()

			// Test that result is valid JSON
			var parsed map[string]interface{}
			err := json.Unmarshal([]byte(result), &parsed)
			if err != nil {
				t.Errorf("String() returned invalid JSON for %s: %v", tt.name, err)
				return
			}

			// Test that required fields are present
			if _, ok := parsed["name"]; !ok {
				t.Errorf("String() missing name field for %s", tt.name)
			}

			if _, ok := parsed["type"]; !ok {
				t.Errorf("String() missing type field for %s", tt.name)
			}

			if _, ok := parsed["required"]; !ok {
				t.Errorf("String() missing required field for %s", tt.name)
			}
		})
	}
}

// TestPromptTemplateArgument_StringJSONStructure tests the exact JSON structure
func TestPromptTemplateArgument_StringJSONStructure(t *testing.T) {
	tests := []struct {
		name     string
		arg      *PromptTemplateArgument
		expected string
	}{
		{
			name: "slice argument",
			arg: &PromptTemplateArgument{
				Name:     "test",
				Argument: []string{"item1", "item2"},
				Required: true,
			},
			expected: `{"name":"test","type":"[]string","required":true,"items":{"type":"[]string"}}`,
		},
		{
			name: "string argument",
			arg: &PromptTemplateArgument{
				Name:     "message",
				Argument: "hello",
				Required: false,
			},
			expected: `{"name":"message","type":"string","required":false}`,
		},
		{
			name: "nil argument",
			arg: &PromptTemplateArgument{
				Name:     "optional",
				Argument: nil,
				Required: false,
			},
			expected: `{"name":"optional","type":null,"required":false}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.arg.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}

			// Also verify it's valid JSON
			var parsed map[string]interface{}
			err := json.Unmarshal([]byte(result), &parsed)
			if err != nil {
				t.Errorf("String() returned invalid JSON: %v", err)
			}
		})
	}
}

// TestPromptTemplateArgument_StringEmptyResult tests cases that should return empty string
func TestPromptTemplateArgument_StringEmptyResult(t *testing.T) {
	// This test would be useful if there were cases where JSON marshaling fails
	// Currently, the method only returns empty string on JSON marshaling error
	// which is hard to trigger in normal circumstances

	// Test with a valid argument to ensure it doesn't return empty string
	arg := &PromptTemplateArgument{
		Name:     "test",
		Argument: "valid",
		Required: true,
	}

	result := arg.String()
	if result == "" {
		t.Errorf("String() returned empty string for valid argument")
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(result), &parsed)
	if err != nil {
		t.Errorf("String() returned invalid JSON: %v", err)
	}
}

// Benchmark tests for performance measurement
func BenchmarkPromptTemplateArgument_String_String(b *testing.B) {
	arg := &PromptTemplateArgument{
		Name:     "message",
		Argument: "hello world",
		Required: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = arg.String()
	}
}

func BenchmarkPromptTemplateArgument_String_Slice(b *testing.B) {
	arg := &PromptTemplateArgument{
		Name:     "items",
		Argument: []string{"item1", "item2", "item3"},
		Required: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = arg.String()
	}
}

func BenchmarkPromptTemplateArgument_String_Map(b *testing.B) {
	arg := &PromptTemplateArgument{
		Name:     "config",
		Argument: map[string]interface{}{"key": "value", "number": 42},
		Required: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = arg.String()
	}
}
