package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/rosstaco/vexdoc-mcp-go/pkg/api"
)

// mockTool is a simple tool for testing
type mockTool struct {
	name        string
	description string
}

func (m *mockTool) Name() string {
	return m.name
}

func (m *mockTool) Description() string {
	return m.description
}

func (m *mockTool) InputSchema() *api.JSONSchema {
	return &api.JSONSchema{
		Type: "object",
		Properties: map[string]*api.JSONSchema{
			"test": {
				Type:        "string",
				Description: "Test parameter",
			},
		},
		Required: []string{"test"},
	}
}

func (m *mockTool) Execute(ctx context.Context, args map[string]interface{}) (*api.ToolResult, error) {
	return &api.ToolResult{
		Content: []api.Content{
			{
				Type: "text",
				Text: "Test result",
			},
		},
	}, nil
}

func TestNewServer(t *testing.T) {
	server := NewServer()
	if server == nil {
		t.Fatal("NewServer returned nil")
	}
	if server.name != ServerName {
		t.Errorf("Expected name %s, got %s", ServerName, server.name)
	}
	if server.version != ServerVersion {
		t.Errorf("Expected version %s, got %s", ServerVersion, server.version)
	}
}

func TestRegisterTool(t *testing.T) {
	server := NewServer()
	tool := &mockTool{
		name:        "test-tool",
		description: "A test tool",
	}

	err := server.RegisterTool(tool)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	// Try registering the same tool again
	err = server.RegisterTool(tool)
	if err == nil {
		t.Error("Expected error when registering duplicate tool, got nil")
	}
}

func TestListTools(t *testing.T) {
	server := NewServer()
	tool1 := &mockTool{name: "tool1", description: "Tool 1"}
	tool2 := &mockTool{name: "tool2", description: "Tool 2"}

	server.RegisterTool(tool1)
	server.RegisterTool(tool2)

	tools := server.ListTools()
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
}

func TestHandleInitialize(t *testing.T) {
	server := NewServer()
	params := api.InitializeRequest{
		ProtocolVersion: ProtocolVersion,
		ClientInfo: api.ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}

	paramsJSON, _ := json.Marshal(params)
	req := &api.Request{
		JSONRPC: JSONRPCVersion,
		ID:      1,
		Method:  MethodInitialize,
		Params:  paramsJSON,
	}

	resp := server.handleInitialize(req)
	if resp.Error != nil {
		t.Errorf("Initialize failed: %v", resp.Error)
	}

	result, ok := resp.Result.(api.InitializeResult)
	if !ok {
		t.Fatal("Result is not InitializeResult")
	}

	if result.ServerInfo.Name != ServerName {
		t.Errorf("Expected server name %s, got %s", ServerName, result.ServerInfo.Name)
	}
}

func TestHandleToolsList(t *testing.T) {
	server := NewServer()
	tool := &mockTool{name: "test-tool", description: "Test"}
	server.RegisterTool(tool)

	req := &api.Request{
		JSONRPC: JSONRPCVersion,
		ID:      1,
		Method:  MethodToolsList,
	}

	resp := server.handleToolsList(req)
	if resp.Error != nil {
		t.Errorf("Tools list failed: %v", resp.Error)
	}

	// Need to convert result through JSON to get proper type
	resultJSON, _ := json.Marshal(resp.Result)
	var result api.ToolsListResult
	json.Unmarshal(resultJSON, &result)

	if len(result.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(result.Tools))
	}
}

func TestHandleToolsCall(t *testing.T) {
	server := NewServer()
	tool := &mockTool{name: "test-tool", description: "Test"}
	server.RegisterTool(tool)

	params := api.ToolCallParams{
		Name: "test-tool",
		Arguments: map[string]interface{}{
			"test": "value",
		},
	}

	paramsJSON, _ := json.Marshal(params)
	req := &api.Request{
		JSONRPC: JSONRPCVersion,
		ID:      1,
		Method:  MethodToolsCall,
		Params:  paramsJSON,
	}

	resp := server.handleToolsCall(context.Background(), req)
	if resp.Error != nil {
		t.Errorf("Tool call failed: %v", resp.Error)
	}
}

func TestHandleToolsCallNotFound(t *testing.T) {
	server := NewServer()

	params := api.ToolCallParams{
		Name: "nonexistent-tool",
		Arguments: map[string]interface{}{
			"test": "value",
		},
	}

	paramsJSON, _ := json.Marshal(params)
	req := &api.Request{
		JSONRPC: JSONRPCVersion,
		ID:      1,
		Method:  MethodToolsCall,
		Params:  paramsJSON,
	}

	resp := server.handleToolsCall(context.Background(), req)
	if resp.Error == nil {
		t.Error("Expected error for nonexistent tool")
	}
	if resp.Error.Code != MethodNotFound {
		t.Errorf("Expected error code %d, got %d", MethodNotFound, resp.Error.Code)
	}
}

func TestHandleMethodNotFound(t *testing.T) {
	server := NewServer()
	req := &api.Request{
		JSONRPC: JSONRPCVersion,
		ID:      1,
		Method:  "nonexistent/method",
	}

	resp := server.handleRequest(context.Background(), req)
	if resp.Error == nil {
		t.Error("Expected error for nonexistent method")
	}
	if resp.Error.Code != MethodNotFound {
		t.Errorf("Expected error code %d, got %d", MethodNotFound, resp.Error.Code)
	}
}
