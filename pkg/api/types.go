package api

import "encoding/json"

// Request represents an MCP JSON-RPC request
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents an MCP JSON-RPC response
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Notification represents an MCP JSON-RPC notification (no response expected)
type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Error represents an MCP error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ToolResult represents the result of tool execution
type ToolResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content represents a piece of content in a tool result
type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ToolInfo contains metadata about a tool
type ToolInfo struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema *JSONSchema `json:"inputSchema"`
}

// JSONSchema represents a JSON Schema for tool parameters
type JSONSchema struct {
	Type                 string                 `json:"type"`
	Properties           map[string]*JSONSchema `json:"properties,omitempty"`
	Required             []string               `json:"required,omitempty"`
	Items                *JSONSchema            `json:"items,omitempty"`
	Enum                 []string               `json:"enum,omitempty"`
	Description          string                 `json:"description,omitempty"`
	AdditionalProperties interface{}            `json:"additionalProperties,omitempty"`
}

// CreateOptions contains options for VEX statement creation
type CreateOptions struct {
	Product       string
	Vulnerability string
	Status        string
	Justification string
	Author        string
	ID            string
}

// MergeOptions contains options for VEX document merging
type MergeOptions struct {
	Documents [][]byte
	OutputID  string
	Author    string
}

// ServerCapabilities represents the capabilities advertised by the server
type ServerCapabilities struct {
	Tools struct {
		ListChanged bool `json:"listChanged,omitempty"`
	} `json:"tools,omitempty"`
}

// ClientCapabilities represents the capabilities advertised by the client
type ClientCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
}

// InitializeRequest represents the initialize method parameters
type InitializeRequest struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

// InitializeResult represents the initialize method result
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

// ClientInfo contains information about the MCP client
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerInfo contains information about the MCP server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ToolsListResult represents the result of tools/list
type ToolsListResult struct {
	Tools []ToolInfo `json:"tools"`
}

// ToolCallParams represents the parameters for tools/call
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}
