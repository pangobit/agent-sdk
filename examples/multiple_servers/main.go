package main

import (
	"fmt"
	"net/http"

	agentsdk "github.com/pangobit/agent-sdk/pkg"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Not the agent server!")
	})
	go http.ListenAndServe(":8080", nil)

	agentServer := agentsdk.NewDefaultServer()
	agentServer.ListenAndServe(":8081")
}
