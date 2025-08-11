package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
)

func main() {
	_ = rpc.Register("")
	rpc.HandleHTTP()

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Println("serving on port 1234")
	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatal(err)
	}

}
