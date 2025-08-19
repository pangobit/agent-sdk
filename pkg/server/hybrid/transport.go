package hybrid

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"strings"
	"sync"
	"time"

	"github.com/pangobit/agent-sdk/pkg/server/jsonrpc"
)

// HybridTransportOpts defines options for configuring the hybrid transport
type HybridTransportOpts func(*HybridTransport)

// WithPath sets the base path for the HTTP endpoints
func WithPath(path string) HybridTransportOpts {
	return func(t *HybridTransport) {
		t.basePath = path
	}
}

// WithReadDeadline sets the read deadline for connections
func WithReadDeadline(d time.Duration) HybridTransportOpts {
	return func(t *HybridTransport) {
		t.readDeadline = d
	}
}

// WithWriteDeadline sets the write deadline for connections
func WithWriteDeadline(d time.Duration) HybridTransportOpts {
	return func(t *HybridTransport) {
		t.writeDeadline = d
	}
}

// HybridTransport provides HTTP endpoints for tool discovery and execution
// while internally using JSON-RPC for method calls
type HybridTransport struct {
	basePath       string
	readDeadline   time.Duration
	writeDeadline  time.Duration
	rpcServer      *rpc.Server
	messageFraming *jsonrpc.MessageFramingService
	toolRegistry   *ToolRegistry
}

// ToolInfo represents information about a registered tool
type ToolInfo struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Parameters  map[string]ParameterInfo `json:"parameters"`
	Returns     string                   `json:"returns"`
	Example     string                   `json:"example,omitempty"`
}

// ParameterInfo represents information about a tool parameter
type ParameterInfo struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// NewHybridTransport creates a new hybrid transport
func NewHybridTransport(opts ...HybridTransportOpts) *HybridTransport {
	t := &HybridTransport{
		basePath:       "/agents/api/v1/",
		readDeadline:   10 * time.Second,
		writeDeadline:  10 * time.Second,
		rpcServer:      rpc.NewServer(),
		messageFraming: jsonrpc.NewMessageFramingService(),
		toolRegistry:   NewToolRegistry(),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// RegisterWithSchema registers a service with the RPC server
func (t *HybridTransport) RegisterWithSchema(rcvr any) error {
	// Register type mappings with the message framing service
	if err := t.messageFraming.RegisterServiceWithTypeMapping(rcvr); err != nil {
		return fmt.Errorf("failed to register service type mappings: %w", err)
	}

	// Register the service with the underlying RPC server
	if err := t.rpcServer.Register(rcvr); err != nil {
		return fmt.Errorf("failed to register service with RPC server: %w", err)
	}

	return nil
}

// GetToolRegistry returns the tool registry for external tool registration
func (t *HybridTransport) GetToolRegistry() *ToolRegistry {
	return t.toolRegistry
}

// GetMessageFramingService returns the message framing service
func (t *HybridTransport) GetMessageFramingService() *jsonrpc.MessageFramingService {
	return t.messageFraming
}

// ListenAndServe starts the server with hybrid transport
func (t *HybridTransport) ListenAndServe(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	defer listener.Close()

	return t.serve(listener)
}

// serve handles incoming connections
func (t *HybridTransport) serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func(conn net.Conn) {
			t.rpcServer.ServeCodec(&hybridCodec{
				decoder:        json.NewDecoder(conn),
				encoder:        json.NewEncoder(conn),
				closer:         conn,
				pending:        make(map[uint64]*json.RawMessage),
				framingService: t.messageFraming,
			})
		}(conn)
	}
}

// HTTPHandler provides HTTP endpoints for tool discovery and execution
func (t *HybridTransport) HTTPHandler() http.Handler {
	mux := http.NewServeMux()

	// Handle the base path
	basePath := strings.TrimSuffix(t.basePath, "/")

	// Create a sub-mux for the API routes
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Agent Tools API - Use /tools for discovery, /execute for execution"))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found!"))
		}
	})
	apiMux.HandleFunc("/tools", t.toolsHandler)
	apiMux.HandleFunc("/execute", t.executeHandler)

	// Mount the API routes under the base path
	mux.Handle(basePath+"/", http.StripPrefix(basePath, apiMux))

	return mux
}

// hybridCodec implements the rpc.ServerCodec interface for hybrid transport
type hybridCodec struct {
	decoder        *json.Decoder
	encoder        *json.Encoder
	closer         interface{ Close() error }
	pending        map[uint64]*json.RawMessage
	framingService *jsonrpc.MessageFramingService

	// Request state
	request struct {
		JSONRPC string           `json:"jsonrpc"`
		Method  string           `json:"method"`
		Params  *json.RawMessage `json:"params"`
		ID      *json.RawMessage `json:"id"`
	}

	mutex sync.Mutex
	seq   uint64
}

func (c *hybridCodec) ReadRequestHeader(r *rpc.Request) error {
	c.request.Method = ""
	c.request.Params = nil
	c.request.ID = nil

	if err := c.decoder.Decode(&c.request); err != nil {
		return err
	}

	if c.request.Method == "" {
		return fmt.Errorf("missing method")
	}

	if c.request.JSONRPC != "2.0" {
		return fmt.Errorf("invalid JSON-RPC version")
	}

	r.ServiceMethod = c.request.Method

	// Handle ID mapping
	c.mutex.Lock()
	c.seq++
	c.pending[c.seq] = c.request.ID
	c.request.ID = nil
	r.Seq = c.seq
	c.mutex.Unlock()

	return nil
}

func (c *hybridCodec) ReadRequestBody(body any) error {
	if body == nil {
		return nil
	}

	// If we have parameters and a framing service, process them
	if c.request.Params != nil && c.framingService != nil {
		// Use framing service to validate and transform messages
		transformedMessages, err := c.framingService.ValidateAndTransformMessages(c.request.Method, *c.request.Params)
		if err != nil {
			return fmt.Errorf("message validation failed: %w", err)
		}

		// Use transformed messages for unmarshaling
		var params [1]any
		params[0] = body
		return json.Unmarshal(transformedMessages, &params)
	}

	// Fallback to original behavior
	var params [1]any
	if c.request.Params != nil {
		params[0] = body
	} else {
		params[0] = nil
	}

	return json.Unmarshal(*c.request.Params, &params)
}

func (c *hybridCodec) WriteResponse(r *rpc.Response, body any) error {
	c.mutex.Lock()
	b, ok := c.pending[r.Seq]
	if !ok {
		c.mutex.Unlock()
		return fmt.Errorf("invalid sequence number")
	}
	delete(c.pending, r.Seq)
	c.mutex.Unlock()

	if b == nil {
		null := json.RawMessage([]byte("null"))
		b = &null
	}

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      b,
	}

	if r.Error == "" {
		response["result"] = body
	} else {
		response["error"] = map[string]interface{}{
			"code":    -32603, // Internal error
			"message": r.Error,
		}
	}

	return c.encoder.Encode(response)
}

func (c *hybridCodec) Close() error {
	return c.closer.Close()
}
