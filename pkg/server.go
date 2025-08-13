package agentsdk

import (
	"context"
	"time"

	"github.com/pangobit/agent-sdk/pkg/server"
	"github.com/pangobit/agent-sdk/pkg/server/http"
	"github.com/pangobit/agent-sdk/pkg/server/sqlite"
)

// NewServer creates a new server with the default HTTP transport and sqlite based repository.
// Pass in [dbopts] to use a custom database/repo pair via [server.WithDB]
func NewServer(context context.Context, dbopts server.ServerOpts, opts ...server.ServerOpts) *server.Server {
	if dbopts == nil {
		opts = append(opts, sqlite.WithDefaultDB())
	} else {
		opts = append(opts, dbopts)
	}

	t := http.NewHTTPTransport(http.WithReadDeadline(10*time.Second), http.WithWriteDeadline(10*time.Second))
	opts = append(opts, server.WithTransport(t))
	return server.NewServer(opts...)
}
