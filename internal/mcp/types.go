package mcp

import "github.com/rosstaco/vexdoc-mcp-go/pkg/api"

// Standard JSON-RPC error codes
const (
	// ParseError - Invalid JSON was received by the server
	ParseError = -32700
	// InvalidRequest - The JSON sent is not a valid Request object
	InvalidRequest = -32600
	// MethodNotFound - The method does not exist / is not available
	MethodNotFound = -32601
	// InvalidParams - Invalid method parameter(s)
	InvalidParams = -32602
	// InternalError - Internal JSON-RPC error
	InternalError = -32603
)

// MCP Protocol Constants
const (
	JSONRPCVersion  = "2.0"
	ProtocolVersion = "2024-11-05"
	ServerName      = "vexdoc-mcp-server"
	ServerVersion   = "0.1.0"
)

// MCP Method Names
const (
	MethodInitialize = "initialize"
	MethodToolsList  = "tools/list"
	MethodToolsCall  = "tools/call"
)

// NewErrorResponse creates a standard error response
func NewErrorResponse(id interface{}, code int, message string, data interface{}) *api.Response {
	return &api.Response{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Error: &api.Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// NewSuccessResponse creates a standard success response
func NewSuccessResponse(id interface{}, result interface{}) *api.Response {
	return &api.Response{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result:  result,
	}
}
