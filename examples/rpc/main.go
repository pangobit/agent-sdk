package main

import (
	"fmt"
	"log"

	"github.com/pangobit/agent-sdk/pkg/jsonrpc"
	"github.com/pangobit/agent-sdk/pkg/server"
	jsonrpctransport "github.com/pangobit/agent-sdk/pkg/server/jsonrpc"
)

type HelloService struct{}

func (h *HelloService) Hello(name string, reply *string) error {
	r := "Hello, " + name
	*reply = r
	fmt.Println("reply", r)
	return nil
}

func main() {
	// Create JSON-RPC transport
	transport := jsonrpctransport.NewJSONRPCTransport()

	// Register the service with the transport
	transport.Register(new(HelloService))

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

	startClient()
}

func startClient() {
	rpcClient, err := jsonrpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal(err)
	}
	defer rpcClient.Close()

	var reply string
	err = rpcClient.Call("HelloService.Hello", "world", &reply)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(reply)
}
