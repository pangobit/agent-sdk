package http

import "github.com/pangobit/agent-sdk/pkg/server"

type HttpServer struct {
	conn server.Connection
}

func NewHttpServer(conn server.Connection) *HttpServer {
	return &HttpServer{conn: conn}
}
