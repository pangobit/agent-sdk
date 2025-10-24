package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/pangobit/agent-sdk/pkg/apigen"
)

// UserProfileRequest represents a request to get user profile
type UserProfileRequest struct {
	UserID         string `json:"user_id" bind:"path=user_id,required"`
	IncludeDetails bool   `json:"include_details" bind:"query=include_details"`
	Format         string `json:"format" bind:"query=format" validate:"omitempty,oneof=json xml"`
}

// UpdateProfileRequest represents a request to update user profile
type UpdateProfileRequest struct {
	UserID   string                 `json:"user_id" bind:"path=user_id,required"`
	Name     string                 `json:"name" bind:"json" validate:"required,min=2,max=100"`
	Email    string                 `json:"email" bind:"json" validate:"required,email"`
	Metadata map[string]interface{} `json:"metadata" bind:"json"`
}

// GetUserProfile handles GET /api/users/{userId}/profile
// This endpoint retrieves a user's profile information.
// Parameters:
//   - userId: The unique identifier for the user (from path, required)
//   - includeDetails: Whether to include detailed profile information (from query)
//   - format: Response format, defaults to json (from query, validates json|xml)
func GetUserProfile(w http.ResponseWriter, r *http.Request, req UserProfileRequest) error {
	// Simulate fetching user profile
	response := map[string]interface{}{
		"userId":  req.UserID,
		"name":    "John Doe",
		"email":   "john@example.com",
		"details": req.IncludeDetails,
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

// UpdateUserProfile handles PUT /api/users/{userId}/profile
// This endpoint updates a user's profile information.
// Parameters:
//   - userId: The unique identifier for the user (from path, required)
//   - name: The user's full name (from JSON body, required, 2-100 chars)
//   - email: The user's email address (from JSON body, required, valid email)
//   - metadata: Additional metadata for the profile (from JSON body)
func UpdateUserProfile(w http.ResponseWriter, r *http.Request, req UpdateProfileRequest) error {
	// Simulate updating user profile
	response := map[string]interface{}{
		"success": true,
		"userId":  req.UserID,
		"updated": []string{"name", "email", "metadata"},
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

// SearchUsers handles POST /api/users/search
// This endpoint searches for users based on criteria.
// Parameters:
//   - query: Search query string
//   - limit: Maximum number of results to return
//   - filters: Additional search filters
func SearchUsers(w http.ResponseWriter, r *http.Request, query string, limit int, filters map[string]interface{}) error {
	// Simulate user search
	response := map[string]interface{}{
		"query":   query,
		"limit":   limit,
		"filters": filters,
		"results": []map[string]interface{}{
			{"userId": "user1", "name": "Alice"},
			{"userId": "user2", "name": "Bob"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

func main() {
	// First, demonstrate the API generation
	fmt.Println("=== API Generation Demo ===")

	// Generate API description from this file using new v2.0 API
	parser := apigen.NewParser()
	transformer := apigen.NewTransformer(apigen.NewTypeRegistry())

	methods, err := parser.ParseSingleFile("examples/apigen_web_server/main.go")
	if err != nil {
		log.Fatalf("Failed to parse methods: %v", err)
	}

	filtered := apigen.FilterByList(methods, []string{"GetUserProfile", "UpdateUserProfile", "SearchUsers"})
	enriched, err := transformer.Transform(filtered)
	if err != nil {
		log.Fatalf("Failed to transform methods: %v", err)
	}

	desc, err := apigen.NewDescription("UserServiceAPI", enriched)
	if err != nil {
		log.Fatalf("Failed to create API description: %v", err)
	}

	// Pretty print the JSON output
	jsonData, err := json.MarshalIndent(desc, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal to JSON: %v", err)
	}

	fmt.Println("Generated API Description:")
	fmt.Println(string(jsonData))
	fmt.Println()

	// Note: In a real implementation, these handlers would be registered with
	// an HTTP router. For this demo, we're just showing the API generation.
	fmt.Println("=== Web Server Demo (Handlers would be registered with router) ===")
	fmt.Println("This example demonstrates how apigen extracts business logic parameters")
	fmt.Println("from HTTP handlers while automatically excluding http.Request and http.ResponseWriter")
	fmt.Println("parameters, making it perfect for generating API documentation for web services.")
}
