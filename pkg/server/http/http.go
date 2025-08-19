// http provides adapters and functionality to serve the server over HTTP
// It wires up the stdlib's http functionality to the server package's connection type
package http

import (
	"net/http"
	"strings"
	"time"
)

type HTTPTransport struct {
	readDeadline  time.Duration
	writeDeadline time.Duration
	basePath      string
	toolHandler   http.Handler // Tool handler injected via options
}

type HTTPTransportOpts func(*HTTPTransport)

func NewHTTPTransport(opts ...HTTPTransportOpts) *HTTPTransport {
	t := &HTTPTransport{}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func WithReadDeadline(d time.Duration) HTTPTransportOpts {
	return func(t *HTTPTransport) {
		t.readDeadline = d
	}
}

func WithWriteDeadline(d time.Duration) HTTPTransportOpts {
	return func(t *HTTPTransport) {
		t.writeDeadline = d
	}
}

func WithPath(path string) HTTPTransportOpts {
	return func(t *HTTPTransport) {
		t.basePath = path
	}
}

// WithToolHandler sets the tool handler for tool-related endpoints
func WithToolHandler(handler http.Handler) HTTPTransportOpts {
	return func(t *HTTPTransport) {
		t.toolHandler = handler
	}
}

func (s *HTTPTransport) ListenAndServe(addr string) error {
	httpSrv := &http.Server{
		Addr:    addr,
		Handler: s.HTTPHandler(),
	}
	return httpSrv.ListenAndServe()
}

func (s *HTTPTransport) HTTPHandler() http.Handler {
	baseMux := http.NewServeMux()

	subroutes := http.NewServeMux()
	subroutes.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Tool Service API - Use /tools for discovery"))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found!"))
		}
	})

	// If we have a tool handler, use it for tool-related routes
	if s.toolHandler != nil {
		subroutes.Handle("/tools", s.toolHandler)
	}

	strippedHandler := http.StripPrefix(strings.TrimSuffix(s.basePath, "/"), subroutes)
	baseMux.Handle(s.basePath, strippedHandler)

	return baseMux
}
