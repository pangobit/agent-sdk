package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/pangobit/mcp/pkg/jsonrpc"
)

type HelloService struct{}

func (h *HelloService) Hello(name string, reply *string) error {
	r := "Hello, " + name
	*reply = r
	fmt.Println("reply", r)
	return nil
}

func main() {
	server := jsonrpc.NewServer()
	server.Register(new(HelloService))

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		fmt.Println("closing listener")
		listener.Close()
		os.Exit(0)
	}()

	log.Println("serving on port 1234")
	go server.Serve(listener)
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
