package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pangobit/agent-sdk/pkg/jsonrpc"
	"github.com/pangobit/agent-sdk/pkg/server"
	jsonrpctransport "github.com/pangobit/agent-sdk/pkg/server/jsonrpc"
)

// Simple service with basic parameter
type HelloService struct{}

func (h *HelloService) Hello(name string, reply *string) error {
	r := "Hello, " + name
	*reply = r
	fmt.Println("reply", r)
	return nil
}

// Enhanced service with structured parameters
type UserService struct{}

type CreateUserParams struct {
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

func (u *UserService) CreateUser(params CreateUserParams, reply *User) error {
	user := &User{
		ID:    fmt.Sprintf("user_%d", len(params.Name)+len(params.Email)), // Simple ID generation
		Name:  params.Name,
		Email: params.Email,
		Age:   params.Age,
	}
	*reply = *user
	fmt.Printf("Created user: %+v\n", user)
	return nil
}

type SearchUsersParams struct {
	Query           string   `json:"query"`
	Filters         []string `json:"filters"`
	Limit           int      `json:"limit"`
	IncludeInactive bool     `json:"include_inactive"`
}

type SearchResult struct {
	Users []User `json:"users"`
	Total int    `json:"total"`
}

func (u *UserService) SearchUsers(params SearchUsersParams, reply *SearchResult) error {
	// Simulate search results
	users := []User{
		{ID: "user_1", Name: "John Doe", Email: "john@example.com", Age: 30},
		{ID: "user_2", Name: "Jane Smith", Email: "jane@example.com", Age: 25},
	}

	*reply = SearchResult{
		Users: users,
		Total: len(users),
	}
	fmt.Printf("Search results for '%s': %d users found\n", params.Query, len(users))
	return nil
}

func main() {
	// Create JSON-RPC transport with message framing capabilities
	transport := jsonrpctransport.NewMessageFramingTransport()

	// Register services with automatic type mapping generation
	transport.RegisterWithSchema(new(HelloService))
	transport.RegisterWithSchema(new(UserService))

	// Create server with JSON-RPC transport
	srv := server.NewServer(
		server.WithTransport(transport),
	)

	log.Println("serving JSON-RPC on port 1234")
	go func() {
		if err := srv.ListenAndServe(":1234"); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait a moment for server to start
	time.Sleep(100 * time.Millisecond)

	startClient()
}

func startClient() {
	rpcClient, err := jsonrpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal(err)
	}
	defer rpcClient.Close()

	// Test 1: Simple parameter (backward compatibility)
	log.Println("=== Test 1: Simple Parameter ===")
	var reply1 string
	err = rpcClient.Call("HelloService.Hello", "world", &reply1)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		log.Printf("Result: %s", reply1)
	}

	// Test 2: Structured parameters (generic JSON)
	log.Println("\n=== Test 2: Structured Parameters ===")
	createUserParams := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	var reply2 map[string]any
	err = rpcClient.Call("UserService.CreateUser", createUserParams, &reply2)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		log.Printf("Created user: %+v", reply2)
	}

	// Test 3: Complex parameters with arrays and booleans
	log.Println("\n=== Test 3: Complex Parameters ===")
	searchParams := map[string]any{
		"query":            "john",
		"filters":          []string{"active", "verified"},
		"limit":            10,
		"include_inactive": false,
	}

	var reply3 map[string]any
	err = rpcClient.Call("UserService.SearchUsers", searchParams, &reply3)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		log.Printf("Search results: %+v", reply3)
	}

	// Test 4: Invalid parameters (should fail type validation)
	log.Println("\n=== Test 4: Invalid Parameters (Type Validation Test) ===")
	invalidParams := map[string]any{
		"name":  123,             // Wrong type (should be string)
		"email": "invalid-email", // Valid string
		"age":   "not-a-number",  // Wrong type (should be number)
	}

	var reply4 map[string]any
	err = rpcClient.Call("UserService.CreateUser", invalidParams, &reply4)
	if err != nil {
		log.Printf("Expected type validation error: %v", err)
	} else {
		log.Printf("Unexpected success: %+v", reply4)
	}
}
