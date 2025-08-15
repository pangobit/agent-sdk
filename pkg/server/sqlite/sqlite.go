package sqlite

import (
	"database/sql"
	"encoding/json"

	"github.com/pangobit/agent-sdk/pkg/server"
	_ "modernc.org/sqlite"
)

const defaultDBPath = "agent.db"

type SqliteRepository struct {
	db *sql.DB
}

func newSqliteRepository(db *sql.DB) *SqliteRepository {
	return &SqliteRepository{db: db}
}

func WithDefaultDB() server.ServerOpts {
	db, err := sql.Open("sqlite", defaultDBPath)
	if err != nil {
		panic(err)
	}
	return WithDB(db)
}

func WithDB(db *sql.DB) server.ServerOpts {
	repo := newSqliteRepository(db)
	return server.WithToolRepository(repo)
}

const toolSchema = `
CREATE TABLE IF NOT EXISTS tools (
    name TEXT PRIMARY KEY,
    description TEXT,
    schema TEXT
)
`

const getToolQuery = `
SELECT name, description, schema FROM tools WHERE name = ?
`

const getAllToolsQuery = `
SELECT name, description, schema FROM tools
`

const createToolQuery = `
INSERT INTO tools (name, description, schema) VALUES (?, ?, ?)
`

const deleteToolQuery = `
DELETE FROM tools WHERE name = ?
`

const updateToolQuery = `
UPDATE tools SET description = ?, schema = ? WHERE name = ?
`

func (r *SqliteRepository) Init() error {
	_, err := r.db.Exec(toolSchema)
	return err
}

func (r *SqliteRepository) GetTool(name string) (server.Tool, error) {
	row := r.db.QueryRow(getToolQuery, name)
	var tool server.Tool
	var schemaData []byte
	err := row.Scan(&tool.Name, &tool.Description, &schemaData)
	if err != nil {
		return server.Tool{}, err
	}
	err = json.Unmarshal(schemaData, &tool.Schema)
	if err != nil {
		return server.Tool{}, err
	}
	return tool, nil
}

func (r *SqliteRepository) GetAllTools() []server.Tool {
	row := r.db.QueryRow(getAllToolsQuery)
	var tool server.Tool
	var schemaData []byte
	err := row.Scan(&tool.Name, &tool.Description, &schemaData)
	if err != nil {
		return nil
	}
	err = json.Unmarshal(schemaData, &tool.Schema)
	if err != nil {
		return nil
	}
	return []server.Tool{tool}
}

func (r *SqliteRepository) CreateTool(tool server.Tool) error {
	schemaData, err := json.Marshal(tool.Schema)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(createToolQuery, tool.Name, tool.Description, schemaData)
	return err
}

func (r *SqliteRepository) DeleteTool(name string) error {
	_, err := r.db.Exec(deleteToolQuery, name)
	return err
}

func (r *SqliteRepository) UpdateTool(name string, tool server.Tool) error {
	schemaData, err := json.Marshal(tool.Schema)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(updateToolQuery, tool.Description, schemaData, name)
	return err
}
