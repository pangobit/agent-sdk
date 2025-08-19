package main

import (
	"fmt"

	agentsdk "github.com/pangobit/agent-sdk/pkg"
	"github.com/pangobit/agent-sdk/pkg/server"
	"github.com/pangobit/agent-sdk/pkg/server/http"
	_ "modernc.org/sqlite"
)

func main() {
	pathOverride := "/my/path/"
	opts := []server.ServerOpts{
		server.WithTransport(
			http.NewHTTPTransport(
				http.WithPath(pathOverride),
			),
		),
	}
	server := agentsdk.NewServer(opts...)

	fmt.Println("serving on http://localhost:8080" + pathOverride)
	server.ListenAndServe(":8080")
}
