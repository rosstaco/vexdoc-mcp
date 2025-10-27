// main.go - MCP Protocol Structure Test
package main

import (
	"encoding/json"
	"fmt"
)

// MCP Request structure
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCP Response structure
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCP Notification structure
type MCPNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Tool-specific structures
type Tool struct {
	Name        string      `json:"name"`
	Title       string      `json:"title,omitempty"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type ToolCallParams struct {
	Name      string      `json:"name"`
	Arguments interface{} `json:"arguments,omitempty"`
}

type ToolResult struct {
	Content []interface{} `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

func main() {
	fmt.Println("MCP Protocol Structure Test")

	// Test 1: Parse a tools/list request
	testToolsListRequest()

	// Test 2: Create a tools/list response
	testToolsListResponse()

	// Test 3: Parse a tools/call request
	testToolsCallRequest()

	// Test 4: Create a tools/call response
	testToolsCallResponse()

	// Test 5: Test error handling
	testErrorResponse()
}

func testToolsListRequest() {
	fmt.Println("\n=== Testing tools/list Request ===")

	reqJSON := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/list",
		"params": {}
	}`

	var req MCPRequest
	if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
		fmt.Printf("Failed to parse request: %v\n", err)
		return
	}

	fmt.Printf("✓ Parsed MCP tools/list request: Method=%s, ID=%v\n", req.Method, req.ID)
}

func testToolsListResponse() {
	fmt.Println("\n=== Testing tools/list Response ===")

	// Create mock VEX tools
	tools := []Tool{
		{
			Name:        "vex-create",
			Title:       "Create VEX Document",
			Description: "Creates a new VEX document with specified vulnerability and product information",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"vulnerability": map[string]interface{}{
						"type":        "string",
						"description": "Vulnerability identifier (e.g., CVE-2023-1234)",
					},
					"product": map[string]interface{}{
						"type":        "string",
						"description": "Product identifier",
					},
					"status": map[string]interface{}{
						"type": "string",
						"enum": []string{"not_affected", "affected", "fixed", "under_investigation"},
					},
				},
				"required": []string{"vulnerability", "product", "status"},
			},
		},
		{
			Name:        "vex-merge",
			Title:       "Merge VEX Documents",
			Description: "Merges multiple VEX documents into a single document",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"documents": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
						"description": "Array of VEX document paths to merge",
					},
				},
				"required": []string{"documents"},
			},
		},
	}

	result := ToolsListResult{Tools: tools}
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result:  result,
	}

	respJSON, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Printf("Failed to create response: %v\n", err)
		return
	}

	fmt.Printf("✓ Created tools/list response:\n%s\n", string(respJSON))
}

func testToolsCallRequest() {
	fmt.Println("\n=== Testing tools/call Request ===")

	reqJSON := `{
		"jsonrpc": "2.0",
		"id": 2,
		"method": "tools/call",
		"params": {
			"name": "vex-create",
			"arguments": {
				"vulnerability": "CVE-2023-1234",
				"product": "myapp@1.0.0",
				"status": "not_affected",
				"justification": "component_not_present"
			}
		}
	}`

	var req MCPRequest
	if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
		fmt.Printf("Failed to parse request: %v\n", err)
		return
	}

	// Parse params as ToolCallParams
	paramsJSON, _ := json.Marshal(req.Params)
	var params ToolCallParams
	if err := json.Unmarshal(paramsJSON, &params); err != nil {
		fmt.Printf("Failed to parse params: %v\n", err)
		return
	}

	fmt.Printf("✓ Parsed tools/call request: Tool=%s\n", params.Name)
	fmt.Printf("  Arguments: %+v\n", params.Arguments)
}

func testToolsCallResponse() {
	fmt.Println("\n=== Testing tools/call Response ===")

	// Mock VEX document creation result
	vexDocument := map[string]interface{}{
		"@context": "https://openvex.dev/ns/v0.2.5",
		"@id":      "generated-vex-001",
		"author":   "vexdoc-mcp-server",
		"version":  1,
		"statements": []interface{}{
			map[string]interface{}{
				"vulnerability": map[string]interface{}{
					"name": "CVE-2023-1234",
				},
				"products": []interface{}{
					map[string]interface{}{
						"@id": "myapp@1.0.0",
					},
				},
				"status":        "not_affected",
				"justification": "component_not_present",
			},
		},
	}

	result := ToolResult{
		Content: []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": "Successfully created VEX document",
			},
			map[string]interface{}{
				"type": "text",
				"text": fmt.Sprintf("VEX Document:\n```json\n%s\n```", mustMarshalJSON(vexDocument)),
			},
		},
		IsError: false,
	}

	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      2,
		Result:  result,
	}

	respJSON, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Printf("Failed to create response: %v\n", err)
		return
	}

	fmt.Printf("✓ Created tools/call response (truncated for readability)\n")
	fmt.Printf("  Response structure: %d content items\n", len(result.Content))
	fmt.Printf("  Full response available (%d bytes)\n", len(respJSON))
}

func testErrorResponse() {
	fmt.Println("\n=== Testing Error Response ===")

	errorResp := MCPResponse{
		JSONRPC: "2.0",
		ID:      3,
		Error: &MCPError{
			Code:    -32602,
			Message: "Invalid params",
			Data:    map[string]interface{}{"reason": "Missing required parameter 'vulnerability'"},
		},
	}

	respJSON, err := json.MarshalIndent(errorResp, "", "  ")
	if err != nil {
		fmt.Printf("Failed to create error response: %v\n", err)
		return
	}

	fmt.Printf("✓ Created error response:\n%s\n", string(respJSON))
}

func mustMarshalJSON(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}
