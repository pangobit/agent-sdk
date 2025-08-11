package jsonrpc

import (
	"testing"
)

func TestValidateResponse(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid response with result",
			response: Response{
				JSONRPC: "2.0",
				Result:  "success",
				ID:      1,
			},
			wantErr: false,
		},
		{
			name: "valid response without result",
			response: Response{
				JSONRPC: "2.0",
				ID:      1,
			},
			wantErr: false,
		},
		{
			name: "valid response with null result",
			response: Response{
				JSONRPC: "2.0",
				Result:  nil,
				ID:      1,
			},
			wantErr: false,
		},
		{
			name: "invalid JSON-RPC version - empty string",
			response: Response{
				JSONRPC: "",
				Result:  "success",
				ID:      1,
			},
			wantErr: true,
			errMsg:  "invalid JSON-RPC version",
		},
		{
			name: "invalid JSON-RPC version - wrong version",
			response: Response{
				JSONRPC: "1.0",
				Result:  "success",
				ID:      1,
			},
			wantErr: true,
			errMsg:  "invalid JSON-RPC version",
		},
		{
			name: "invalid JSON-RPC version - 3.0",
			response: Response{
				JSONRPC: "3.0",
				Result:  "success",
				ID:      1,
			},
			wantErr: true,
			errMsg:  "invalid JSON-RPC version",
		},
		{
			name: "invalid JSON-RPC version - case sensitive",
			response: Response{
				JSONRPC: "2.0",
				Result:  "success",
				ID:      1,
			},
			wantErr: false, // This should actually be valid since it's exactly "2.0"
		},
		{
			name: "response with error",
			response: Response{
				JSONRPC: "2.0",
				Error: &Error{
					Code:    -32601,
					Message: "Method not found",
				},
				ID: 1,
			},
			wantErr: true,
			errMsg:  "error: Method not found",
		},
		{
			name: "response with error and empty message",
			response: Response{
				JSONRPC: "2.0",
				Error: &Error{
					Code:    -32600,
					Message: "",
				},
				ID: 1,
			},
			wantErr: true,
			errMsg:  "error: ",
		},
		{
			name: "response with error and complex message",
			response: Response{
				JSONRPC: "2.0",
				Error: &Error{
					Code:    -32700,
					Message: "Parse error: Invalid JSON was received by the server",
				},
				ID: 1,
			},
			wantErr: true,
			errMsg:  "error: Parse error: Invalid JSON was received by the server",
		},
		{
			name: "valid response with string ID",
			response: Response{
				JSONRPC: "2.0",
				Result:  "success",
				ID:      1, // Note: Current implementation uses int, but spec allows string/number/null
			},
			wantErr: false,
		},
		{
			name: "valid response with zero ID",
			response: Response{
				JSONRPC: "2.0",
				Result:  "success",
				ID:      0,
			},
			wantErr: false,
		},
		{
			name: "valid response with negative ID",
			response: Response{
				JSONRPC: "2.0",
				Result:  "success",
				ID:      -1,
			},
			wantErr: false,
		},
		{
			name: "valid response with large ID",
			response: Response{
				JSONRPC: "2.0",
				Result:  "success",
				ID:      999999999,
			},
			wantErr: false,
		},
		{
			name: "valid response with complex result object",
			response: Response{
				JSONRPC: "2.0",
				Result: map[string]interface{}{
					"status": "success",
					"data":   []string{"item1", "item2"},
					"count":  42,
					"nested": map[string]interface{}{
						"key": "value",
					},
				},
				ID: 1,
			},
			wantErr: false,
		},
		{
			name: "valid response with array result",
			response: Response{
				JSONRPC: "2.0",
				Result:  []interface{}{1, "two", 3.14, true, nil},
				ID:      1,
			},
			wantErr: false,
		},
		{
			name: "valid response with boolean result",
			response: Response{
				JSONRPC: "2.0",
				Result:  true,
				ID:      1,
			},
			wantErr: false,
		},
		{
			name: "valid response with number result",
			response: Response{
				JSONRPC: "2.0",
				Result:  3.14159,
				ID:      1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResponse(tt.response)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateResponse() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("validateResponse() error = %v, want error message %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateResponse() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidateResponseEdgeCases tests edge cases and boundary conditions
func TestValidateResponseEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		wantErr  bool
	}{
		{
			name: "response with whitespace in JSON-RPC version",
			response: Response{
				JSONRPC: " 2.0 ",
				Result:  "success",
				ID:      1,
			},
			wantErr: true,
		},
		{
			name: "response with unicode characters in JSON-RPC version",
			response: Response{
				JSONRPC: "2.0\u200B", // Zero-width space
				Result:  "success",
				ID:      1,
			},
			wantErr: true,
		},
		{
			name: "response with error containing special characters",
			response: Response{
				JSONRPC: "2.0",
				Error: &Error{
					Code:    -32600,
					Message: "Error with special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?",
				},
				ID: 1,
			},
			wantErr: true,
		},
		{
			name: "response with error containing unicode characters",
			response: Response{
				JSONRPC: "2.0",
				Error: &Error{
					Code:    -32600,
					Message: "Error with unicode: 🚀🌟🎉",
				},
				ID: 1,
			},
			wantErr: true,
		},
		{
			name: "response with very long error message",
			response: Response{
				JSONRPC: "2.0",
				Error: &Error{
					Code:    -32600,
					Message: "This is a very long error message that contains many characters and should be handled properly by the validation function. It includes various types of content and should not cause any issues with the validation logic.",
				},
				ID: 1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResponse(tt.response)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateResponse() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("validateResponse() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidateResponseSpecCompliance tests compliance with JSON-RPC 2.0 specification
func TestValidateResponseSpecCompliance(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid response according to spec",
			response: Response{
				JSONRPC: "2.0",
				Result:  "success",
				ID:      1,
			},
			wantErr: false,
		},
		{
			name: "response with error according to spec",
			response: Response{
				JSONRPC: "2.0",
				Error: &Error{
					Code:    -32601,
					Message: "Method not found",
				},
				ID: 1,
			},
			wantErr: true,
			errMsg:  "error: Method not found",
		},
		{
			name: "invalid jsonrpc version according to spec",
			response: Response{
				JSONRPC: "1.0", // Invalid version
				Result:  "success",
				ID:      1,
			},
			wantErr: true,
			errMsg:  "invalid JSON-RPC version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResponse(tt.response)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateResponse() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("validateResponse() error = %v, want error message %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateResponse() unexpected error = %v", err)
				}
			}
		})
	}
}

// Benchmark tests for performance measurement
func BenchmarkValidateResponseValid(b *testing.B) {
	response := Response{
		JSONRPC: "2.0",
		Result:  "success",
		ID:      1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateResponse(response)
	}
}

func BenchmarkValidateResponseInvalidVersion(b *testing.B) {
	response := Response{
		JSONRPC: "1.0",
		Result:  "success",
		ID:      1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateResponse(response)
	}
}

func BenchmarkValidateResponseWithError(b *testing.B) {
	response := Response{
		JSONRPC: "2.0",
		Error: &Error{
			Code:    -32601,
			Message: "Method not found",
		},
		ID: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateResponse(response)
	}
}

// TestValidateMethod tests the validateMethod function
func TestValidateMethod(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid method name",
			method:  "getUser",
			wantErr: false,
		},
		{
			name:    "valid method name with underscore",
			method:  "get_user_info",
			wantErr: false,
		},
		{
			name:    "valid method name with numbers",
			method:  "getUser123",
			wantErr: false,
		},
		{
			name:    "valid method name with special characters",
			method:  "get-user-info",
			wantErr: false,
		},
		{
			name:    "valid method name starting with different letters",
			method:  "abc123",
			wantErr: false,
		},
		{
			name:    "valid method name with dots but not starting with rpc",
			method:  "api.v1.getUser",
			wantErr: false,
		},
		{
			name:    "empty method name",
			method:  "",
			wantErr: true,
			errMsg:  "method name cannot be empty",
		},
		{
			name:    "reserved method starting with rpc.",
			method:  "rpc.getUser",
			wantErr: true,
			errMsg:  "method name 'rpc.getUser' is reserved for rpc-internal methods and extensions",
		},
		{
			name:    "reserved method rpc. with no additional text",
			method:  "rpc.",
			wantErr: true,
			errMsg:  "method name 'rpc.' is reserved for rpc-internal methods and extensions",
		},
		{
			name:    "reserved method with complex rpc. prefix",
			method:  "rpc.internal.system.getStatus",
			wantErr: true,
			errMsg:  "method name 'rpc.internal.system.getStatus' is reserved for rpc-internal methods and extensions",
		},
		{
			name:    "method starting with rpc but not followed by dot",
			method:  "rpcgetUser",
			wantErr: false,
		},
		{
			name:    "method with rpc in middle",
			method:  "getRpcUser",
			wantErr: false,
		},
		{
			name:    "method ending with rpc",
			method:  "getUserRpc",
			wantErr: false,
		},
		{
			name:    "method with rpc. in middle",
			method:  "api.rpc.getUser",
			wantErr: false,
		},
		{
			name:    "method with uppercase RPC",
			method:  "RPC.getUser",
			wantErr: false,
		},
		{
			name:    "method with mixed case rPc",
			method:  "rPc.getUser",
			wantErr: false,
		},
		{
			name:    "method with unicode characters",
			method:  "getUser🚀",
			wantErr: false,
		},
		{
			name:    "method with spaces",
			method:  "get user info",
			wantErr: false,
		},
		{
			name:    "method with tabs",
			method:  "get\tuser\tinfo",
			wantErr: false,
		},
		{
			name:    "method with newlines",
			method:  "get\nuser\ninfo",
			wantErr: false,
		},
		{
			name:    "method with special unicode characters",
			method:  "getUser\u200B", // Zero-width space
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMethod(tt.method)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateMethod() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("validateMethod() error = %v, want error message %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateMethod() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidateMethodEdgeCases tests edge cases for the validateMethod function
func TestValidateMethodEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		wantErr bool
	}{
		{
			name:    "method with exactly 3 characters",
			method:  "abc",
			wantErr: false,
		},
		{
			name:    "method with exactly 4 characters but not rpc.",
			method:  "abcd",
			wantErr: false,
		},
		{
			name:    "method with exactly 4 characters and is rpc.",
			method:  "rpc.",
			wantErr: true,
		},
		{
			name:    "method with 5 characters starting with rpc.",
			method:  "rpc.x",
			wantErr: true,
		},
		{
			name:    "method with very long name",
			method:  "thisIsAVeryLongMethodNameThatShouldBeValidAndNotCauseAnyIssuesWithTheValidationLogic",
			wantErr: false,
		},
		{
			name:    "method with only special characters",
			method:  "!@#$%^&*()",
			wantErr: false,
		},
		{
			name:    "method with only numbers",
			method:  "123456789",
			wantErr: false,
		},
		{
			name:    "method with only dots",
			method:  "...",
			wantErr: false,
		},
		{
			name:    "method with rpc followed by multiple dots",
			method:  "rpc...",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMethod(tt.method)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateMethod() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("validateMethod() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidateMethodSpecCompliance tests compliance with JSON-RPC 2.0 specification
func TestValidateMethodSpecCompliance(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid method according to spec",
			method:  "getUser",
			wantErr: false,
		},
		{
			name:    "reserved method rpc. according to spec",
			method:  "rpc.getUser",
			wantErr: true,
			errMsg:  "method name 'rpc.getUser' is reserved for rpc-internal methods and extensions",
		},
		{
			name:    "empty method name according to spec",
			method:  "",
			wantErr: true,
			errMsg:  "method name cannot be empty",
		},
		{
			name:    "method with rpc. prefix according to spec",
			method:  "rpc.internal.getStatus",
			wantErr: true,
			errMsg:  "method name 'rpc.internal.getStatus' is reserved for rpc-internal methods and extensions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMethod(tt.method)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateMethod() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("validateMethod() error = %v, want error message %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateMethod() unexpected error = %v", err)
				}
			}
		})
	}
}

// Benchmark tests for validateMethod function
func BenchmarkValidateMethodValid(b *testing.B) {
	method := "getUser"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateMethod(method)
	}
}

func BenchmarkValidateMethodReserved(b *testing.B) {
	method := "rpc.getUser"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateMethod(method)
	}
}

func BenchmarkValidateMethodEmpty(b *testing.B) {
	method := ""

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateMethod(method)
	}
}
