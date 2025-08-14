package main

import (
	"context"
	"fmt"

	agentsdk "github.com/pangobit/agent-sdk/pkg"
	_ "modernc.org/sqlite"
)

func main() {
	server := agentsdk.NewDefaultServer(context.Background())

	//fmt.Println("server", server)
	// Agent server endpoints will be mounted at the domain root path by default
	fmt.Println("serving on http://localhost:8080")
	server.ListenAndServe(":8080")
}
