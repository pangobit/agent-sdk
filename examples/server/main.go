package main

import (
	"context"

	agentsdk "github.com/pangobit/agent-sdk/pkg"
	_ "modernc.org/sqlite"
)

func main() {
	server := agentsdk.NewServer(context.Background(), nil)

	server.ListenAndServe(":8080")
}
