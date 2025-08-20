// http provides adapters and functionality to serve the server over HTTP
// It wires up the stdlib's http functionality to the server package's connection type
package http

import (
	"net/http"
	"strings"
	"time"
)

// HTTPTransport implements the [server.Transport interface
type HTTPTransport struct {
	readDeadline  time.Duration
	writeDeadline time.Duration
	basePath      string
	toolHandler   http.Handler
	methodHandler http.Handler
}

type HTTPTransportOpts func(*HTTPTransport)

// NewHTTPTransport creates a new HTTP transport and applies the given options
func NewHTTPTransport(opts ...HTTPTransportOpts) *HTTPTransport {
	t := &HTTPTransport{}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// WithReadDeadline sets the read deadline for the HTTP transport
// the read deadline is the maximum amount of time a request can take to be read before
// the request is timed out
func WithReadDeadline(d time.Duration) HTTPTransportOpts {
	return func(t *HTTPTransport) {
		t.readDeadline = d
	}
}

// WithWriteDeadline sets the write deadline for the HTTP transport
// the write deadline is the maximum amount of time a response can take to be written before
// the response is timed out
func WithWriteDeadline(d time.Duration) HTTPTransportOpts {
	return func(t *HTTPTransport) {
		t.writeDeadline = d
	}
}

// WithPath sets the base path for the HTTP transport
// the base path is the path that the HTTP transport will be mounted at
// E.g., if the base path is "/my/path", the HTTP transport will be mounted
// at "http://host:8080/my/path" (assuming the server is running on port 8080)
func WithPath(path string) HTTPTransportOpts {
	return func(t *HTTPTransport) {
		t.basePath = path
	}
}

// WithToolHandler sets the tool http handler for tool-related endpoints
func WithToolHandler(handler http.Handler) HTTPTransportOpts {
	return func(t *HTTPTransport) {
		t.toolHandler = handler
	}
}

// WithMethodHandler sets the method execution handler
func WithMethodHandler(handler http.Handler) HTTPTransportOpts {
	return func(t *HTTPTransport) {
		t.methodHandler = handler
	}
}

// ListenAndServe starts the HTTP transport and listens for incoming requests
// the addr is the address to listen on
// E.g., if the addr is ":8080", the HTTP transport will listen on port 8080
func (s *HTTPTransport) ListenAndServe(addr string) error {
	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      s.HTTPHandler(),
		ReadTimeout:  s.readDeadline,
		WriteTimeout: s.writeDeadline,
	}

	return httpSrv.ListenAndServe()
}

// HTTPHandler returns the HTTP handler for the HTTP transport
// the handler is a mux that handles the base path and agent-related endpoints
// the base path is the path that the HTTP transport will be mounted at
func (s *HTTPTransport) HTTPHandler() http.Handler {
	// Create subroutes with common handlers
	subroutes := s.createSubroutes()

	// If no base path is set, return the subroutes directly
	if s.basePath == "" {
		return subroutes
	}

	// If base path is set, use the full routing logic
	baseMux := http.NewServeMux()

	// Handle the base path with proper prefix stripping
	basePath := strings.TrimSuffix(s.basePath, "/")
	if basePath == "" {
		basePath = "/"
	}

	strippedHandler := http.StripPrefix(basePath, subroutes)
	baseMux.Handle(basePath+"/", strippedHandler)

	return baseMux
}

// createSubroutes creates the subroutes with common handlers
func (s *HTTPTransport) createSubroutes() http.Handler {
	subroutes := http.NewServeMux()

	// Root handler
	subroutes.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Tool Service API - Use /tools for discovery, /execute for method calls"))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found!"))
		}
	})

	// Tool discovery handler
	if s.toolHandler != nil {
		subroutes.Handle("/tools", s.toolHandler)
	}

	// Method execution handler
	if s.methodHandler != nil {
		subroutes.Handle("/execute", s.methodHandler)
	}

	return subroutes
}
