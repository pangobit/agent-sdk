// http provides adapters and functionality to serve the server over HTTP
// It wires up the stdlib's http functionality to the server package's connection type
package http

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	stdhttp "net/http"
	"time"
)

type HTTPTransport struct {
	readDeadline  time.Duration
	writeDeadline time.Duration
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

type httpConnection struct {
	io.ReadWriteCloser
}

func (c *httpConnection) LocalAddr() net.Addr {
	return nil
}

func (c *httpConnection) RemoteAddr() net.Addr {
	return nil
}

func (c *httpConnection) SetDeadline(t time.Time) error {
	return nil
}

func (c *httpConnection) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *httpConnection) SetWriteDeadline(t time.Time) error {
	return nil
}

func (s *HTTPTransport) ListenAndServe(addr string) error {
	httpSrv := &stdhttp.Server{
		Addr:    addr,
		Handler: s.HTTPHandler(),
	}
	return httpSrv.ListenAndServe()
}

func (s *HTTPTransport) ServeConn(listener net.Listener) error {
	srv := &stdhttp.Server{
		Handler: s.HTTPHandler(),
	}
	return srv.Serve(listener)
}

func (s *HTTPTransport) HTTPHandler() stdhttp.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hello, World!")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	return mux
}

type httpListener struct {
	net.Conn
}

func (l *httpListener) Accept() (net.Conn, error) {
	if l.Conn == nil {
		return nil, errors.New("no more connections")
	}
	c := l.Conn
	l.Conn = nil
	return c, nil
}

func (l *httpListener) Close() error {
	return nil
}

func (l *httpListener) Addr() net.Addr {
	return l.Conn.LocalAddr()
}
