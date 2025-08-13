// jsonrpc provides adapters and functionality to serve the server over JSON-RPC
// It wires up the jsonrpc package's functionality to the server package's connection type
package jsonrpc

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/pangobit/agent-sdk/pkg/jsonrpc"
)

type JSONRPCTransport struct {
	readDeadline  time.Duration
	writeDeadline time.Duration
	server        *jsonrpc.Server
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

// HTTPHandler provides an HTTP endpoint for JSON-RPC over HTTP
func (t *JSONRPCTransport) HTTPHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set content type for JSON-RPC
		w.Header().Set("Content-Type", "application/json")

		// For HTTP transport, we'll handle the JSON-RPC manually
		// since we can't access the private serverCodec
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Simple echo response for now
		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req["id"],
			"result":  "Hello from JSON-RPC over HTTP",
		}

		json.NewEncoder(w).Encode(response)
	})

	return mux
}
