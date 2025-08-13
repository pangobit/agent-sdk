package agentsdk

import (
	"time"

	"github.com/pangobit/agent-sdk/pkg/server"
	"github.com/pangobit/agent-sdk/pkg/server/http"
)

// NewServer creates a new server with the default HTTP transport.
func NewServer(opts ...server.ServerOpts) *server.Server {
	t := http.NewHTTPTransport(http.WithReadDeadline(10*time.Second), http.WithWriteDeadline(10*time.Second))
	opts = append(opts, server.WithTransport(t))
	return server.NewServer(opts...)
}
