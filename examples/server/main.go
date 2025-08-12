package main

import (
	"database/sql"
	"log"

	"github.com/pangobit/agent-sdk/pkg/server"
	"github.com/pangobit/agent-sdk/pkg/server/sqlite"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "mcp.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_ = server.NewServer(sqlite.WithSqliteRepository(db))
}
