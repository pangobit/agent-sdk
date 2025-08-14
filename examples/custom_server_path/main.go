package main

import (
	"context"
	"fmt"

	agentsdk "github.com/pangobit/agent-sdk/pkg"
	"github.com/pangobit/agent-sdk/pkg/server"
	_ "modernc.org/sqlite"
)

const pathOverride = "/my/path"

func main() {
	server := agentsdk.NewServer(context.Background(), nil, server.WithPath(pathOverride))

	// This will now listen on http://localhost:8080/my/path
	fmt.Println("serving on http://localhost:8080/my/path")
	server.ListenAndServe(":8080")
}
