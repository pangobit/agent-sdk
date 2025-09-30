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
	UserID   string `json:"userId"`
	IncludeDetails bool   `json:"includeDetails"`
}

// UpdateProfileRequest represents a request to update user profile
type UpdateProfileRequest struct {
	UserID   string                 `json:"userId"`
	Name     string                 `json:"name"`
	Email    string                 `json:"email"`
	Metadata map[string]interface{} `json:"metadata"`
}

// GetUserProfile handles GET /api/users/{userId}/profile
// This endpoint retrieves a user's profile information.
// Parameters:
//   - userId: The unique identifier for the user
//   - includeDetails: Whether to include detailed profile information
func GetUserProfile(w http.ResponseWriter, r *http.Request, req UserProfileRequest) error {
	// Simulate fetching user profile
	response := map[string]interface{}{
		"userId":   req.UserID,
		"name":     "John Doe",
		"email":    "john@example.com",
		"details":  req.IncludeDetails,
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

// UpdateUserProfile handles PUT /api/users/{userId}/profile
// This endpoint updates a user's profile information.
// Parameters:
//   - userId: The unique identifier for the user
//   - name: The user's full name
//   - email: The user's email address
//   - metadata: Additional metadata for the profile
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

	// Generate API description from this file
	config := apigen.WithMethodList("GetUserProfile", "UpdateUserProfile", "SearchUsers").
		SetAPIName("UserServiceAPI")

	desc, err := apigen.GenerateFromFile("examples/apigen_web_server/main.go", config)
	if err != nil {
		log.Fatalf("Failed to generate API description: %v", err)
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