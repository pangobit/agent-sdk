// http provides adapters and functionality to serve the server over HTTP
// It wires up the stdlib's http functionality to the server package's connection type
package http

import (
	"errors"
	"io"
	"net"
	stdhttp "net/http"
	"time"

	"github.com/pangobit/agent-sdk/pkg/server"
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
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		conn.SetReadDeadline(time.Now().Add(s.readDeadline))
		conn.SetWriteDeadline(time.Now().Add(s.writeDeadline))
		go s.ServeConn(conn)
	}
}

func (s *HTTPTransport) ServeConn(conn server.Connection) error {
	c := &httpConnection{conn}
	srv := &stdhttp.Server{
		Handler: s.HTTPHandler(),
	}
	l := &httpListener{Conn: c}
	return srv.Serve(l)
}

func (s *HTTPTransport) HTTPHandler() stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {

	})
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
	return l.Conn.Close()
}

func (l *httpListener) Addr() net.Addr {
	return l.Conn.LocalAddr()
}
