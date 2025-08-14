package main

import (
	"context"
	"fmt"

	agentsdk "github.com/pangobit/agent-sdk/pkg"
	"github.com/pangobit/agent-sdk/pkg/server"
	"github.com/pangobit/agent-sdk/pkg/server/http"
	_ "modernc.org/sqlite"
)

const pathOverride = "/my/path"

func main() {
	opts := []server.ServerOpts{
		server.WithTransportOpts(
			http.GetWithPathOption(pathOverride),
		),
	}
	server := agentsdk.NewServer(context.Background(), nil, opts...)

	// Agent server endpoints will be mounted at http://localhost:8080/my/path
	fmt.Println("serving on http://localhost:8080/my/path")
	server.ListenAndServe(":8080")
}
