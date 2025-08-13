package main

import (
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/pangobit/agent-sdk/pkg/server"
	"github.com/pangobit/agent-sdk/pkg/server/sqlite"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "agent.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	_ = server.NewServer(sqlite.WithSqliteRepository(db))

}
