package main

import (
	"testing"

	"github.com/pangobit/agent-sdk/pkg/apigen"
)

func TestApplyFiltering(t *testing.T) {
	methods := []apigen.RawMethod{
		{Name: "HandleUser"},
		{Name: "ProcessData"},
		{Name: "ValidateInput"},
		{Name: "HandleRequest"},
	}

	tests := []struct {
		name           string
		methodList     string
		prefix         string
		suffix         string
		contains       string
		expectedLength int
		expectedNames  []string
	}{
		{
			name:           "no filtering",
			methodList:     "",
			prefix:         "",
			suffix:         "",
			contains:       "",
			expectedLength: 4,
			expectedNames:  []string{"HandleUser", "ProcessData", "ValidateInput", "HandleRequest"},
		},
		{
			name:           "filter by prefix",
			methodList:     "",
			prefix:         "Handle",
			suffix:         "",
			contains:       "",
			expectedLength: 2,
			expectedNames:  []string{"HandleUser", "HandleRequest"},
		},
		{
			name:           "filter by suffix",
			methodList:     "",
			prefix:         "",
			suffix:         "Input",
			contains:       "",
			expectedLength: 1,
			expectedNames:  []string{"ValidateInput"},
		},
		{
			name:           "filter by contains",
			methodList:     "",
			prefix:         "",
			suffix:         "",
			contains:       "Data",
			expectedLength: 1,
			expectedNames:  []string{"ProcessData"},
		},
		{
			name:           "filter by method list",
			methodList:     "HandleUser,ValidateInput",
			prefix:         "",
			suffix:         "",
			contains:       "",
			expectedLength: 2,
			expectedNames:  []string{"HandleUser", "ValidateInput"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyFiltering(methods, tt.methodList, tt.prefix, tt.suffix, tt.contains)

			if len(result) != tt.expectedLength {
				t.Errorf("expected %d methods, got %d", tt.expectedLength, len(result))
			}

			// Check that expected methods are present
			resultNames := make([]string, len(result))
			for i, method := range result {
				resultNames[i] = method.Name
			}

			for _, expectedName := range tt.expectedNames {
				found := false
				for _, resultName := range resultNames {
					if resultName == expectedName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected method %s not found in result", expectedName)
				}
			}
		})
	}
}

func TestParseCommaSeparated(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single item",
			input:    "item1",
			expected: []string{"item1"},
		},
		{
			name:     "multiple items",
			input:    "item1,item2,item3",
			expected: []string{"item1", "item2", "item3"},
		},
		{
			name:     "items with spaces",
			input:    "item1, item2 , item3",
			expected: []string{"item1", " item2 ", " item3"},
		},
		{
			name:     "quoted items",
			input:    `"item1","item2","item3"`,
			expected: []string{"item1", "item2", "item3"},
		},
		{
			name:     "mixed quoted and unquoted",
			input:    `item1,"item2",item3`,
			expected: []string{"item1", "item2", "item3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCommaSeparated(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d items, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("expected item %d to be %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}

func TestInferPackageName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple package",
			input:    "./pkg/handlers",
			expected: "handlers",
		},
		{
			name:     "nested package",
			input:    "github.com/user/project/pkg/handlers",
			expected: "handlers",
		},
		{
			name:     "single component",
			input:    "main",
			expected: "main",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferPackageName(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}