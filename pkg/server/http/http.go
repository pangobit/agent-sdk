// http provides adapters and functionality to serve the server over HTTP
// It wires up the stdlib's http functionality to the server package's connection type
package http

import (
	"fmt"
	"net/http"
	stdhttp "net/http"
	"time"

	"github.com/pangobit/agent-sdk/pkg/server"
)

type HTTPTransport struct {
	readDeadline  time.Duration
	writeDeadline time.Duration
	basePath      string
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

func GetWithPathOption(path string) server.TransportOpts {
	return func(t server.Transport) server.Transport {
		t.(*HTTPTransport).basePath = path
		return t
	}
}

func (s *HTTPTransport) ListenAndServe(addr string) error {
	httpSrv := &stdhttp.Server{
		Addr:    addr,
		Handler: s.HTTPHandler(),
	}
	return httpSrv.ListenAndServe()
}

func (s *HTTPTransport) HTTPHandler() stdhttp.Handler {
	baseMux := http.NewServeMux()

	subroutes := http.NewServeMux()
	subroutes.HandleFunc("/hello/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hello, World!")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	// Mount the subroutes at the specific base path
	fmt.Println("Mounting subroutes at", s.basePath)
	baseMux.Handle(s.basePath, http.StripPrefix(s.basePath, subroutes))

	return baseMux
}
