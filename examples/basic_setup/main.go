package main

import (
	"fmt"
	"log"

	agentsdk "github.com/pangobit/agent-sdk/pkg"
)

// Simple service for testing
type HelloService struct{}

type HelloRequest struct {
	Name string `json:"name"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

func (h *HelloService) Hello(req HelloRequest, reply *HelloResponse) error {
	reply.Message = "Hello, " + req.Name
	fmt.Printf("Hello called with: %s\n", req.Name)
	return nil
}

type UserService struct{}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func (u *UserService) CreateUser(req CreateUserRequest, reply *User) error {
	user := &User{
		ID:    fmt.Sprintf("user_%d", len(req.Name)+len(req.Email)),
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}
	*reply = *user
	fmt.Printf("CreateUser called with: %+v\n", req)
	return nil
}

func main() {
	// Create default server with tool functionality
	server := agentsdk.NewDefaultServer()

	// Register tools with semantic descriptions
	helloParams := map[string]interface{}{
		"name": map[string]interface{}{
			"type":        "string",
			"description": "The name to greet",
			"required":    true,
		},
	}

	server.RegisterMethod("HelloService", "Hello",
		"Sends a greeting message to the specified name", helloParams)

	// Register user creation tool
	userParams := map[string]interface{}{
		"name": map[string]interface{}{
			"type":        "string",
			"description": "Full name of the user to create",
			"required":    true,
		},
		"email": map[string]interface{}{
			"type":        "string",
			"description": "Email address for the user account",
			"required":    true,
		},
		"age": map[string]interface{}{
			"type":        "number",
			"description": "Age of the user in years",
			"required":    true,
		},
	}

	server.RegisterMethod("UserService", "CreateUser",
		"Creates a new user account with the provided information", userParams)

	fmt.Println("serving API on http://localhost:8080/agents/api/v1/")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /agents/api/v1/tools   - Tool discovery")
	fmt.Println("  POST /agents/api/v1/execute - Tool execution")

	// Start the server
	log.Fatal(server.ListenAndServe(":8080"))
}
