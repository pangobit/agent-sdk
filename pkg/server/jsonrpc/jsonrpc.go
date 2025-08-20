// jsonrpc provides adapters and functionality to serve the server over JSON-RPC
// It wires up the jsonrpc package's functionality to the server package's connection type
package jsonrpc

import (
	"fmt"
	"net"
	"time"

	"github.com/pangobit/agent-sdk/pkg/jsonrpc"
)

type JSONRPCTransport struct {
	readDeadline  time.Duration
	writeDeadline time.Duration
	server        *jsonrpc.Server
	basePath      string
}

type JSONRPCTransportOpts func(*JSONRPCTransport)

func NewJSONRPCTransport(opts ...JSONRPCTransportOpts) *JSONRPCTransport {
	t := &JSONRPCTransport{
		server: jsonrpc.NewServer(),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func WithReadDeadline(d time.Duration) JSONRPCTransportOpts {
	return func(t *JSONRPCTransport) {
		t.readDeadline = d
	}
}

func WithWriteDeadline(d time.Duration) JSONRPCTransportOpts {
	return func(t *JSONRPCTransport) {
		t.writeDeadline = d
	}
}

func WithPath(path string) JSONRPCTransportOpts {
	return func(t *JSONRPCTransport) {
		t.basePath = path
	}
}

func (t *JSONRPCTransport) ListenAndServe(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	defer listener.Close()

	return t.server.Serve(listener)
}

func (t *JSONRPCTransport) Register(rcvr any) error {
	return t.server.Register(rcvr)
}
