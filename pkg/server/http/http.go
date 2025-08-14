// http provides adapters and functionality to serve the server over HTTP
// It wires up the stdlib's http functionality to the server package's connection type
package http

import (
	"fmt"
	"net/http"
	"strings"
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
	httpSrv := &http.Server{
		Addr:    addr,
		Handler: s.HTTPHandler(),
	}
	return httpSrv.ListenAndServe()
}

type apiHandler struct {
}

func (apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Serving request for", r.URL.Path)
	fmt.Println("url is ", r.URL.String())
}

func (s *HTTPTransport) HTTPHandler() http.Handler {
	baseMux := http.NewServeMux()

	subroutes := http.NewServeMux()
	subroutes.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			fmt.Println("Home!")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Home!"))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found!"))
		}
	})
	subroutes.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hello, World!")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})
	strippedHandler := http.StripPrefix(strings.TrimSuffix(s.basePath, "/"), subroutes)
	baseMux.Handle(s.basePath, strippedHandler)

	return baseMux
}
