// parameter_framing.go provides composition APIs to integrate message framing with JSON-RPC transport
package jsonrpc

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"reflect"
	"sync"
	"time"
)

// MessageFramingTransport provides JSON-RPC transport with message framing capabilities
type MessageFramingTransport struct {
	readDeadline   time.Duration
	writeDeadline  time.Duration
	basePath       string
	framingService *MessageFramingService
	server         *rpc.Server
}

// NewMessageFramingTransport creates a new transport with message framing
func NewMessageFramingTransport(opts ...JSONRPCTransportOpts) *MessageFramingTransport {
	t := &MessageFramingTransport{
		framingService: NewMessageFramingService(),
		server:         rpc.NewServer(),
	}
	for _, opt := range opts {
		// Apply options to the embedded JSONRPCTransport fields
		opt(&JSONRPCTransport{
			readDeadline:  t.readDeadline,
			writeDeadline: t.writeDeadline,
			basePath:      t.basePath,
		})
	}
	return t
}

// RegisterWithSchema registers a service and generates type mappings for its methods
func (t *MessageFramingTransport) RegisterWithSchema(rcvr any) error {
	// Register type mappings with the framing service
	if err := t.framingService.RegisterServiceWithTypeMapping(rcvr); err != nil {
		return fmt.Errorf("failed to register service type mappings: %w", err)
	}

	// Register the service with the underlying RPC server
	return t.server.Register(rcvr)
}

// RegisterMethodType manually registers a method with its parameter type
func (t *MessageFramingTransport) RegisterMethodType(methodName string, paramType reflect.Type) {
	t.framingService.RegisterMethodType(methodName, paramType)
}

// ListenAndServe starts the server with message framing
func (t *MessageFramingTransport) ListenAndServe(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	defer listener.Close()

	return t.serveWithFraming(listener)
}

// serveWithFraming serves with custom codec that includes message framing
func (t *MessageFramingTransport) serveWithFraming(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func(conn net.Conn) {
			t.server.ServeCodec(&messageFramingCodec{
				decoder:        json.NewDecoder(conn),
				encoder:        json.NewEncoder(conn),
				closer:         conn,
				pending:        make(map[uint64]*json.RawMessage),
				framingService: t.framingService,
			})
		}(conn)
	}
}

// HTTPHandler provides an HTTP endpoint for JSON-RPC over HTTP with message framing
func (t *MessageFramingTransport) HTTPHandler() http.Handler {
	mux := http.NewServeMux()

	subroutes := http.NewServeMux()
	subroutes.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set content type for JSON-RPC
		w.Header().Set("Content-Type", "application/json")

		// Parse the JSON-RPC request
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Extract method and parameters
		method, ok := req["method"].(string)
		if !ok {
			http.Error(w, "missing method", http.StatusBadRequest)
			return
		}

		// Handle message framing if there are parameters
		if params, exists := req["params"]; exists {
			// Convert params to JSON for processing
			paramsJSON, err := json.Marshal(params)
			if err != nil {
				http.Error(w, "invalid parameters", http.StatusBadRequest)
				return
			}

			// Use framing service to validate and transform messages
			transformedMessages, err := t.framingService.ValidateAndTransformMessages(method, paramsJSON)
			if err != nil {
				// Return JSON-RPC error response
				errorResponse := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      req["id"],
					"error": map[string]interface{}{
						"code":    -32602, // Invalid params
						"message": fmt.Sprintf("Message validation failed: %v", err),
					},
				}
				json.NewEncoder(w).Encode(errorResponse)
				return
			}

			// Update the request with transformed messages
			var transformedMessagesValue interface{}
			if err := json.Unmarshal(transformedMessages, &transformedMessagesValue); err != nil {
				http.Error(w, "message transformation failed", http.StatusInternalServerError)
				return
			}
			req["params"] = transformedMessagesValue
		}

		// Simple echo response for now
		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req["id"],
			"result":  "Hello from JSON-RPC over HTTP with message framing",
		}

		json.NewEncoder(w).Encode(response)
	})

	// Mount the subroutes at the specific base path
	mux.Handle(t.basePath+"/", http.StripPrefix(t.basePath, subroutes))

	return mux
}

// messageFramingCodec implements the rpc.ServerCodec interface with message framing
type messageFramingCodec struct {
	decoder        *json.Decoder
	encoder        *json.Encoder
	closer         interface{ Close() error }
	pending        map[uint64]*json.RawMessage
	framingService *MessageFramingService

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

func (c *messageFramingCodec) ReadRequestHeader(r *rpc.Request) error {
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

func (c *messageFramingCodec) ReadRequestBody(body any) error {
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

func (c *messageFramingCodec) WriteResponse(r *rpc.Response, body any) error {
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

func (c *messageFramingCodec) Close() error {
	return c.closer.Close()
}
