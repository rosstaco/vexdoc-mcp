package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/rosstaco/vexdoc-mcp/pkg/api"
)

// Server represents the MCP server instance
type Server struct {
	name         string
	version      string
	tools        map[string]api.Tool
	capabilities api.ServerCapabilities
	mu           sync.RWMutex
	initialized  bool
}

// NewServer creates a new MCP server instance
func NewServer() *Server {
	return &Server{
		name:    ServerName,
		version: ServerVersion,
		tools:   make(map[string]api.Tool),
		capabilities: api.ServerCapabilities{
			Tools: struct {
				ListChanged bool `json:"listChanged,omitempty"`
			}{
				ListChanged: false,
			},
		},
	}
}

// Start begins the MCP server execution
func (s *Server) Start() error {
	return s.StartWithTransport(context.Background(), NewStdioTransport())
}

// StartWithTransport starts the server with a specific transport
func (s *Server) StartWithTransport(ctx context.Context, transport api.Transport) error {
	defer transport.Close()

	fmt.Fprintln(os.Stderr, "[INFO] MCP Server starting...")
	fmt.Fprintf(os.Stderr, "[INFO] Server: %s v%s\n", s.name, s.version)
	fmt.Fprintf(os.Stderr, "[INFO] Protocol Version: %s\n", ProtocolVersion)

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(os.Stderr, "[INFO] Server shutting down...")
			return ctx.Err()
		default:
			req, err := transport.Read()
			if err != nil {
				if err.Error() == "EOF" {
					fmt.Fprintln(os.Stderr, "[INFO] Connection closed")
					return nil
				}
				fmt.Fprintf(os.Stderr, "[ERROR] Read error: %v\n", err)
				continue
			}

			resp := s.handleRequest(ctx, req)
			if err := transport.Write(resp); err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] Write error: %v\n", err)
				return err
			}
		}
	}
}

// RegisterTool registers a tool with the server
func (s *Server) RegisterTool(tool api.Tool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tools[tool.Name()]; exists {
		return fmt.Errorf("tool %s already registered", tool.Name())
	}

	s.tools[tool.Name()] = tool
	fmt.Fprintf(os.Stderr, "[INFO] Registered tool: %s\n", tool.Name())
	return nil
}

// ListTools returns information about all registered tools
func (s *Server) ListTools() []api.ToolInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]api.ToolInfo, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, api.ToolInfo{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: tool.InputSchema(),
		})
	}
	return tools
}

// Stop stops the server
func (s *Server) Stop() error {
	fmt.Fprintln(os.Stderr, "[INFO] Server stopped")
	return nil
}

// handleRequest routes incoming requests to appropriate handlers
func (s *Server) handleRequest(ctx context.Context, req *api.Request) *api.Response {
	switch req.Method {
	case MethodInitialize:
		return s.handleInitialize(req)
	case MethodToolsList:
		return s.handleToolsList(req)
	case MethodToolsCall:
		return s.handleToolsCall(ctx, req)
	default:
		return NewErrorResponse(req.ID, MethodNotFound,
			fmt.Sprintf("Method not found: %s", req.Method), nil)
	}
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(req *api.Request) *api.Response {
	var params api.InitializeRequest
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return NewErrorResponse(req.ID, InvalidParams,
				"Invalid initialize parameters", err.Error())
		}
	}

	s.mu.Lock()
	s.initialized = true
	s.mu.Unlock()

	result := api.InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities:    s.capabilities,
		ServerInfo: api.ServerInfo{
			Name:    s.name,
			Version: s.version,
		},
	}

	fmt.Fprintf(os.Stderr, "[INFO] Initialized by client: %s v%s\n",
		params.ClientInfo.Name, params.ClientInfo.Version)

	return NewSuccessResponse(req.ID, result)
}

// handleToolsList handles the tools/list request
func (s *Server) handleToolsList(req *api.Request) *api.Response {
	tools := s.ListTools()
	result := api.ToolsListResult{
		Tools: tools,
	}

	fmt.Fprintf(os.Stderr, "[INFO] Listed %d tools\n", len(tools))
	return NewSuccessResponse(req.ID, result)
}

// handleToolsCall handles the tools/call request
func (s *Server) handleToolsCall(ctx context.Context, req *api.Request) *api.Response {
	var params api.ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(req.ID, InvalidParams,
			"Invalid tool call parameters", err.Error())
	}

	s.mu.RLock()
	tool, exists := s.tools[params.Name]
	s.mu.RUnlock()

	if !exists {
		return NewErrorResponse(req.ID, MethodNotFound,
			fmt.Sprintf("Tool not found: %s", params.Name), nil)
	}

	fmt.Fprintf(os.Stderr, "[INFO] Executing tool: %s\n", params.Name)

	result, err := tool.Execute(ctx, params.Arguments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Tool execution failed: %v\n", err)
		return NewErrorResponse(req.ID, InternalError,
			"Tool execution failed", err.Error())
	}

	return NewSuccessResponse(req.ID, result)
}
