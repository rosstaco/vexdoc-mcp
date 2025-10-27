package api

import (
	"context"
)

// MCPServer defines the main server interface
type MCPServer interface {
	Start(ctx context.Context, transport Transport) error
	Stop() error
	RegisterTool(tool Tool) error
	ListTools() []ToolInfo
}

// Transport handles MCP communication
type Transport interface {
	Read() (*Request, error)
	Write(*Response) error
	Close() error
}

// Tool represents an MCP tool
type Tool interface {
	Name() string
	Description() string
	InputSchema() *JSONSchema
	Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
}

// StreamingTool extends Tool with streaming capabilities
type StreamingTool interface {
	Tool
	Stream(ctx context.Context, args map[string]interface{}) (<-chan *ToolResult, error)
}

// VEXClient handles VEX operations
type VEXClient interface {
	CreateStatement(ctx context.Context, opts *CreateOptions) (interface{}, error)
	MergeDocuments(ctx context.Context, opts *MergeOptions) (interface{}, error)
	ValidateDocument(ctx context.Context, doc interface{}) error
}
