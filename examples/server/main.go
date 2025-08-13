package main

import (
	"database/sql"
	"log"

	agentsdk "github.com/pangobit/agent-sdk/pkg"
	"github.com/pangobit/agent-sdk/pkg/server/sqlite"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "agent.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	server := agentsdk.NewServer(sqlite.WithSqliteRepository(db))

	server.ListenAndServe(":8080")
}
