package jsonrpc

import (
	"encoding/json"
	"net"
	"net/rpc"
)

// Server embeds the rpc.Server.
// server is not exported, which inherently limits access to the server API.
// This is a deliberate and opinionated design decision to make the mcp implementation
// easier to understand and maintain.
type Server struct {
	server *rpc.Server
}

// NewServer creates a new json RPC server.
// The default codec is used to ensure that the server is compatible with the json RPC 2.0 spec.
// Again, this is a rigid, but intentional decision.
func NewServer() *Server {
	server := rpc.NewServer()
	return &Server{server: server}
}

// Register registers a new object for use as a service.
// It is a wrapper around the rpc.Server.Register method.
func (s *Server) Register(rcvr any) error {
	return s.server.Register(rcvr)
}

func (s *Server) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go s.server.ServeCodec(&serverCodec{
			decoder: json.NewDecoder(conn),
			encoder: json.NewEncoder(conn),
			closer:  conn,
			pending: make(map[uint64]*json.RawMessage),
		})
	}
}
