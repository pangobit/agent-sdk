package server

type ToolGetter interface {
	GetTool(name string) (Tool, error)
	GetAllTools() []Tool
}

type ToolCreator interface {
	CreateTool(tool Tool) error
}

type ToolDeleter interface {
	DeleteTool(name string) error
}

type ToolUpdater interface {
	UpdateTool(name string, tool Tool) error
}

type ToolRepository interface {
	ToolGetter
	ToolCreator
	ToolDeleter
	ToolUpdater
}
