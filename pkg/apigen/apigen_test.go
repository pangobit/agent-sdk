package apigen_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/pangobit/agent-sdk/pkg/apigen"
)

// TestParser_ParseSingleFile tests parsing of individual Go files
func TestParser_ParseSingleFile(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		wantMethods int
		wantNames   []string
	}{
		{
			name:        "basic_methods",
			filePath:    filepath.Join("testdata", "basic_methods.go"),
			wantMethods: 2,
			wantNames:   []string{"HandleEcho", "HandlePing"},
		},
		{
			name:        "nested_structs",
			filePath:    filepath.Join("testdata", "nested_structs.go"),
			wantMethods: 1,
			wantNames:   []string{"ProcessCompany"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := apigen.NewParser()
			methods, err := parser.ParseSingleFile(tt.filePath)
			if err != nil {
				t.Fatalf("ParseSingleFile() error = %v", err)
			}

			if len(methods) != tt.wantMethods {
				t.Errorf("ParseSingleFile() got %d methods, want %d", len(methods), tt.wantMethods)
			}

			for i, method := range methods {
				if method.Name != tt.wantNames[i] {
					t.Errorf("ParseSingleFile() method %d name = %s, want %s", i, method.Name, tt.wantNames[i])
				}
			}
		})
	}
}

// TestTransformer_Transform tests type transformation
func TestTransformer_Transform(t *testing.T) {
	tests := []struct {
		name             string
		filePath         string
		filterMethod     string
		expectedParam    string
		expectedType     string
		expectFieldsFail bool // true if we expect no Fields (basic types)
		expectElementType bool // true if we expect ElementType to be set (slices)
		expectKeyValueTypes bool // true if we expect KeyType and ValueType to be set (maps)
	}{
		{
			name:             "basic_types",
			filePath:         filepath.Join("testdata", "basic_methods.go"),
			filterMethod:     "HandleEcho",
			expectedParam:    "message",
			expectedType:     "string",
			expectFieldsFail: true, // Basic string type should not have fields
			expectElementType: false,
			expectKeyValueTypes: false,
		},
		{
			name:             "nested_structs",
			filePath:         filepath.Join("testdata", "nested_structs.go"),
			filterMethod:     "ProcessCompany",
			expectedParam:    "company",
			expectedType:     "Company",
			expectFieldsFail: false, // Now should succeed - nested types should be resolved
			expectElementType: false,
			expectKeyValueTypes: false,
		},
		{
			name:             "slice_types",
			filePath:         filepath.Join("testdata", "slice_types.go"),
			filterMethod:     "ProcessUsers",
			expectedParam:    "users",
			expectedType:     "[]User",
			expectFieldsFail: true, // Slice itself shouldn't have fields in Fields map
			expectElementType: true, // But should have ElementType describing the User type
			expectKeyValueTypes: false,
		},
		{
			name:             "map_types",
			filePath:         filepath.Join("testdata", "map_types.go"),
			filterMethod:     "ProcessProfiles",
			expectedParam:    "profiles",
			expectedType:     "map[string]Profile",
			expectFieldsFail: true, // Map itself shouldn't have fields in Fields map
			expectElementType: false,
			expectKeyValueTypes: true, // Should have KeyType and ValueType
		},
		{
			name:             "complex_types",
			filePath:         filepath.Join("testdata", "complex_types.go"),
			filterMethod:     "ProcessTeam",
			expectedParam:    "team",
			expectedType:     "Team",
			expectFieldsFail: false, // Now should succeed - nested types should be resolved
			expectElementType: false,
			expectKeyValueTypes: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := apigen.NewParser()
			transformer := apigen.NewTransformer(parser.GetRegistry())

			methods, err := parser.ParseSingleFile(tt.filePath)
			if err != nil {
				t.Fatalf("failed to parse file: %v", err)
			}

			filtered := apigen.FilterByList(methods, []string{tt.filterMethod})
			if len(filtered) != 1 {
				t.Fatalf("expected 1 method after filtering, got %d", len(filtered))
			}

			enriched, err := transformer.Transform(filtered)
			if err != nil {
				t.Fatalf("Transform() error = %v", err)
			}

			if len(enriched) != 1 {
				t.Fatalf("expected 1 enriched method, got %d", len(enriched))
			}

			method := enriched[0]
			if len(method.Parameters) == 0 {
				t.Fatal("expected at least 1 parameter")
			}

			param := method.Parameters[0]
			if param.Name != tt.expectedParam {
				t.Errorf("expected parameter name %s, got %s", tt.expectedParam, param.Name)
			}

			// Check the type by creating the API description
			desc := apigen.NewDescription("TestAPI", []apigen.EnrichedMethod{method})
			methodDesc := desc.Methods[method.Name]
			paramInfo := methodDesc.Parameters[param.Name]
			
			if paramInfo.Type != tt.expectedType {
				t.Errorf("expected parameter type %s, got %s", tt.expectedType, paramInfo.Type)
			}

			// Check field resolution in the API description
			if tt.expectFieldsFail {
				if len(paramInfo.Fields) > 0 {
					t.Errorf("expected no fields in API description, but got %d fields", len(paramInfo.Fields))
				}
			} else {
				// For types that should be resolved, check that fields are present
				if len(paramInfo.Fields) == 0 {
					t.Errorf("expected fields in API description, but got none")
				}
			}
			
			// Check for ElementType on slices
			if tt.expectElementType {
				if paramInfo.ElementType == nil {
					t.Errorf("expected ElementType to be set for slice type, but it was nil")
				} else if paramInfo.ElementType.Type == "" {
					t.Errorf("expected ElementType to have a type, but it was empty")
				}
			} else {
				if paramInfo.ElementType != nil {
					t.Errorf("expected ElementType to be nil for non-slice type, but it was set")
				}
			}
			
			// Check for KeyType and ValueType on maps
			if tt.expectKeyValueTypes {
				if paramInfo.KeyType == nil {
					t.Errorf("expected KeyType to be set for map type, but it was nil")
				}
				if paramInfo.ValueType == nil {
					t.Errorf("expected ValueType to be set for map type, but it was nil")
				}
			} else {
				if paramInfo.KeyType != nil {
					t.Errorf("expected KeyType to be nil for non-map type, but it was set")
				}
				if paramInfo.ValueType != nil {
					t.Errorf("expected ValueType to be nil for non-map type, but it was set")
				}
			}
		})
	}
}

// TestFilterFunctions tests various filtering strategies
func TestFilterFunctions(t *testing.T) {
	parser := apigen.NewParser()
	methods, err := parser.ParseSingleFile(filepath.Join("testdata", "basic_methods.go"))
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	tests := []struct {
		name     string
		filter   func([]apigen.RawMethod) []apigen.RawMethod
		expected []string
	}{
		{
			name:     "FilterByPrefix_Handle",
			filter:   func(m []apigen.RawMethod) []apigen.RawMethod { return apigen.FilterByPrefix(m, "Handle") },
			expected: []string{"HandleEcho", "HandlePing"},
		},
		{
			name:     "FilterBySuffix_Echo",
			filter:   func(m []apigen.RawMethod) []apigen.RawMethod { return apigen.FilterBySuffix(m, "Echo") },
			expected: []string{"HandleEcho"},
		},
		{
			name:     "FilterByContains_Ping",
			filter:   func(m []apigen.RawMethod) []apigen.RawMethod { return apigen.FilterByContains(m, "Ping") },
			expected: []string{"HandlePing"},
		},
		{
			name:     "FilterByList_specific",
			filter:   func(m []apigen.RawMethod) []apigen.RawMethod { return apigen.FilterByList(m, []string{"HandleEcho"}) },
			expected: []string{"HandleEcho"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := tt.filter(methods)
			if len(filtered) != len(tt.expected) {
				t.Errorf("expected %d methods, got %d", len(tt.expected), len(filtered))
			}

			for i, method := range filtered {
				if method.Name != tt.expected[i] {
					t.Errorf("expected method %s, got %s", tt.expected[i], method.Name)
				}
			}
		})
	}
}

// TestJSONGenerator_Generate tests JSON generation
func TestJSONGenerator_Generate(t *testing.T) {
	parser := apigen.NewParser()
	transformer := apigen.NewTransformer(parser.GetRegistry())
	generator := apigen.NewJSONGenerator()

	methods, err := parser.ParseSingleFile(filepath.Join("testdata", "basic_methods.go"))
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	filtered := apigen.FilterByPrefix(methods, "Handle")
	enriched, err := transformer.Transform(filtered)
	if err != nil {
		t.Fatalf("failed to transform methods: %v", err)
	}

	desc := apigen.NewDescription("TestAPI", enriched)
	content, err := generator.Generate(desc)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify the generated JSON is valid
	var parsed apigen.APIDescription
	if err := json.Unmarshal([]byte(content.Content), &parsed); err != nil {
		t.Fatalf("generated JSON is invalid: %v", err)
	}

	// Verify structure
	if parsed.APIName != "TestAPI" {
		t.Errorf("expected API name 'TestAPI', got '%s'", parsed.APIName)
	}

	if len(parsed.Methods) != 2 {
		t.Errorf("expected 2 methods, got %d", len(parsed.Methods))
	}

	if _, exists := parsed.Methods["HandleEcho"]; !exists {
		t.Error("HandleEcho method not found in generated API")
	}

	if _, exists := parsed.Methods["HandlePing"]; !exists {
		t.Error("HandlePing method not found in generated API")
	}
}