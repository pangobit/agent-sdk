package sqlite

import (
	"database/sql"
	"encoding/json"

	"github.com/pangobit/agent-sdk/pkg/server"
)

type SqliteRepository struct {
	db *sql.DB
}

func NewSqliteRepository(db *sql.DB) *SqliteRepository {
	return &SqliteRepository{db: db}
}

func WithSqliteRepository(db *sql.DB) server.ServerOpts {
	repo := NewSqliteRepository(db)
	return server.WithToolRepository(repo)
}

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
