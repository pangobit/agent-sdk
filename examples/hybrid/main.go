package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pangobit/agent-sdk/pkg/server/hybrid"
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
	// Create hybrid transport with default path /agents/api/v1/
	transport := hybrid.NewHybridTransport(
		hybrid.WithReadDeadline(10*time.Second),
		hybrid.WithWriteDeadline(10*time.Second),
	)

	// Register services with the hybrid transport
	transport.RegisterWithSchema(new(HelloService))
	transport.RegisterWithSchema(new(UserService))

	// Get the tool registry for manual tool registration
	registry := transport.GetToolRegistry()

	// Register tools with semantic descriptions using the builder pattern
	helloParams := hybrid.NewParameterBuilder().
		String("name", "The name to greet").
		Build()

	registry.RegisterMethod("HelloService", "Hello",
		"Sends a greeting message to the specified name", helloParams)

	// Register user creation tool
	userParams := hybrid.NewParameterBuilder().
		String("name", "Full name of the user to create").
		String("email", "Email address for the user account").
		Number("age", "Age of the user in years").
		Build()

	registry.RegisterMethod("UserService", "CreateUser",
		"Creates a new user account with the provided information", userParams)

	fmt.Println("serving hybrid transport on http://localhost:8080/agents/api/v1/")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /agents/api/v1/tools   - Tool discovery")
	fmt.Println("  POST /agents/api/v1/execute - Tool execution")

	// Start the server
	log.Fatal(http.ListenAndServe(":8080", transport.HTTPHandler()))
}
