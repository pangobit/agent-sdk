package jsonrpc

import (
	"encoding/json"
	"net/rpc"
	"os"
)

// Server embeds the rpc.Server.
// server is not exported, which inherently limits access to the server API.
// This is a deliberate and opinionated design decision to make the mcp implementation
// easier to understand and maintain.
type Server struct {
	server *rpc.Server
}

func NewServer() *Server {
	codec := serverCodec{
		decoder: json.NewDecoder(os.Stdin),
		encoder: json.NewEncoder(os.Stdout),
	}
	server := rpc.NewServer()
	server.ServeCodec(&codec)
	return &Server{server: server}
}
