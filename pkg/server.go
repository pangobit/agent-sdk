package agentsdk

import (
	"context"
	"time"

	"github.com/pangobit/agent-sdk/pkg/server"
	"github.com/pangobit/agent-sdk/pkg/server/http"
	"github.com/pangobit/agent-sdk/pkg/server/sqlite"
)

// NewDefaultServer creates a new server with the default HTTP transport and sqlite based repository.
// Unlike the underlying NewServer function, this function does not accept any options.
func NewDefaultServer(context context.Context) *server.Server {
	defaultOpts := []server.ServerOpts{
		server.WithTransport(
			http.NewHTTPTransport(
				http.WithReadDeadline(10*time.Second),
				http.WithWriteDeadline(10*time.Second),
				http.WithPath("/agent/api/"),
			),
		),
	}
	return NewServer(context, nil, defaultOpts...)
}

// NewServer creates a new server with the default HTTP transport and sqlite based repository.
// Pass in [dbopts] to use a custom database/repo pair via [server.WithDB]
func NewServer(context context.Context, dbopts server.ServerOpts, opts ...server.ServerOpts) *server.Server {
	if dbopts == nil {
		opts = append(opts, sqlite.WithDefaultDB())
	} else {
		opts = append(opts, dbopts)
	}

	return server.NewServer(opts...)
}
