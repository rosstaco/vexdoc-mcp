# Phase 7: Testing & Validation
**Story Points**: 8 | **Prerequisites**: [Phase 6](./phase6-streaming.md) | **Next**: [Phase 8](./phase8-deployment.md)

## Overview
Implement comprehensive testing suite covering unit tests, integration tests, performance tests, and compatibility validation to ensure production readiness.

## Objectives
- [ ] Achieve 90%+ test coverage across all modules
- [ ] Validate compatibility with existing MCP clients
- [ ] Ensure output format matches Node.js implementation exactly
- [ ] Validate performance improvements and memory usage
- [ ] Create automated test suite for CI/CD pipeline

## Tasks

### Task 7.1: Unit Testing Suite (3 points)

**Goal**: Comprehensive unit tests for all components with high coverage

**Checklist:**
- [ ] Test all VEX client operations (create, merge, validate)
- [ ] Test MCP protocol handler and transport layers
- [ ] Test tool implementations with various inputs
- [ ] Test error handling and edge cases
- [ ] Test configuration and logging systems

**Code Example - VEX Client Unit Tests:**
```go
// internal/vex/client_test.go
package vex

import (
    "context"
    "encoding/json"
    "testing"
    "time"
    
    "github.com/rosstaco/vexdoc-mcp-go/pkg/api"
    "github.com/rosstaco/vexdoc-mcp-go/internal/logging"
    "github.com/rosstaco/vexdoc-mcp-go/test/helpers"
)

func TestClient_CreateStatement(t *testing.T) {
    tests := []struct {
        name    string
        opts    *api.CreateOptions
        wantErr bool
        validate func(*testing.T, *api.VEXDocument)
    }{
        {
            name: "valid_not_affected_statement",
            opts: &api.CreateOptions{
                Product:       "nginx:1.20",
                Vulnerability: "CVE-2023-1234",
                Status:        "not_affected",
                Justification: "component_not_present",
                Author:        "test-author",
            },
            wantErr: false,
            validate: func(t *testing.T, doc *api.VEXDocument) {
                if doc.ID == "" {
                    t.Error("Document ID should not be empty")
                }
                if len(doc.Statements) != 1 {
                    t.Errorf("Expected 1 statement, got %d", len(doc.Statements))
                }
                if doc.Statements[0].Status != "not_affected" {
                    t.Errorf("Expected status 'not_affected', got %s", doc.Statements[0].Status)
                }
                if doc.Author != "test-author" {
                    t.Errorf("Expected author 'test-author', got %s", doc.Author)
                }
            },
        },
        {
            name: "valid_fixed_statement",
            opts: &api.CreateOptions{
                Product:       "pkg:golang/github.com/gin-gonic/gin@v1.9.1",
                Vulnerability: "CVE-2023-5678",
                Status:        "fixed",
            },
            wantErr: false,
            validate: func(t *testing.T, doc *api.VEXDocument) {
                if doc.Statements[0].Status != "fixed" {
                    t.Errorf("Expected status 'fixed', got %s", doc.Statements[0].Status)
                }
            },
        },
        {
            name: "invalid_status",
            opts: &api.CreateOptions{
                Product:       "nginx:1.20",
                Vulnerability: "CVE-2023-1234",
                Status:        "invalid_status",
            },
            wantErr: true,
        },
        {
            name: "empty_product",
            opts: &api.CreateOptions{
                Product:       "",
                Vulnerability: "CVE-2023-1234",
                Status:        "not_affected",
            },
            wantErr: true,
        },
        {
            name: "invalid_cve_format",
            opts: &api.CreateOptions{
                Product:       "nginx:1.20",
                Vulnerability: "INVALID-CVE",
                Status:        "not_affected",
            },
            wantErr: true,
        },
    }
    
    client := setupTestClient(t)
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            doc, err := client.CreateStatement(context.Background(), tt.opts)
            
            if tt.wantErr {
                if err == nil {
                    t.Error("Expected error but got none")
                }
                return
            }
            
            if err != nil {
                t.Errorf("Unexpected error: %v", err)
                return
            }
            
            if doc == nil {
                t.Error("Expected document but got nil")
                return
            }
            
            // Run custom validation if provided
            if tt.validate != nil {
                tt.validate(t, doc)
            }
            
            // Validate JSON serialization
            jsonBytes, err := json.Marshal(doc)
            if err != nil {
                t.Errorf("Failed to marshal document: %v", err)
            }
            
            var unmarshaled api.VEXDocument
            if err := json.Unmarshal(jsonBytes, &unmarshaled); err != nil {
                t.Errorf("Failed to unmarshal document: %v", err)
            }
        })
    }
}

func TestClient_MergeDocuments(t *testing.T) {
    tests := []struct {
        name      string
        documents []string
        wantErr   bool
        validate  func(*testing.T, *api.VEXDocument, []string)
    }{
        {
            name: "merge_two_simple_documents",
            documents: []string{
                `{
                    "@id": "doc1",
                    "author": "author1",
                    "version": 1,
                    "timestamp": "2023-01-01T00:00:00Z",
                    "statements": [{
                        "vulnerability": {"name": "CVE-2023-1234"},
                        "products": [{"component": {"@id": "product1"}}],
                        "status": "not_affected",
                        "justification": "component_not_present"
                    }]
                }`,
                `{
                    "@id": "doc2", 
                    "author": "author2",
                    "version": 1,
                    "timestamp": "2023-01-02T00:00:00Z",
                    "statements": [{
                        "vulnerability": {"name": "CVE-2023-5678"},
                        "products": [{"component": {"@id": "product2"}}],
                        "status": "fixed"
                    }]
                }`,
            },
            wantErr: false,
            validate: func(t *testing.T, merged *api.VEXDocument, inputs []string) {
                if len(merged.Statements) != 2 {
                    t.Errorf("Expected 2 statements, got %d", len(merged.Statements))
                }
            },
        },
        {
            name: "merge_conflicting_statements",
            documents: []string{
                `{
                    "@id": "doc1",
                    "author": "author1", 
                    "version": 1,
                    "timestamp": "2023-01-01T00:00:00Z",
                    "statements": [{
                        "vulnerability": {"name": "CVE-2023-1234"},
                        "products": [{"component": {"@id": "product1"}}],
                        "status": "affected"
                    }]
                }`,
                `{
                    "@id": "doc2",
                    "author": "author2",
                    "version": 1, 
                    "timestamp": "2023-01-02T00:00:00Z",
                    "statements": [{
                        "vulnerability": {"name": "CVE-2023-1234"},
                        "products": [{"component": {"@id": "product1"}}],
                        "status": "fixed"
                    }]
                }`,
            },
            wantErr: false,
            validate: func(t *testing.T, merged *api.VEXDocument, inputs []string) {
                if len(merged.Statements) != 1 {
                    t.Errorf("Expected 1 merged statement, got %d", len(merged.Statements))
                }
                // Should resolve to "fixed" (higher priority than "affected")
                if merged.Statements[0].Status != "fixed" {
                    t.Errorf("Expected status 'fixed', got %s", merged.Statements[0].Status)
                }
            },
        },
        {
            name: "single_document_error",
            documents: []string{
                `{"@id": "doc1", "statements": []}`,
            },
            wantErr: true,
        },
        {
            name: "invalid_json_document",
            documents: []string{
                `{"@id": "doc1", "statements": []}`,
                `invalid json`,
            },
            wantErr: true,
        },
    }
    
    client := setupTestClient(t)
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            opts := &api.MergeOptions{
                Documents: tt.documents,
                OutputID:  "test-merged-doc",
                Author:    "test-merger",
            }
            
            merged, err := client.MergeDocuments(context.Background(), opts)
            
            if tt.wantErr {
                if err == nil {
                    t.Error("Expected error but got none")
                }
                return
            }
            
            if err != nil {
                t.Errorf("Unexpected error: %v", err)
                return
            }
            
            if merged == nil {
                t.Error("Expected merged document but got nil")
                return
            }
            
            // Run custom validation
            if tt.validate != nil {
                tt.validate(t, merged, tt.documents)
            }
        })
    }
}

func TestClient_StreamMerge(t *testing.T) {
    client := setupTestClient(t)
    
    // Generate test documents
    documents := make([]string, 50)
    for i := 0; i < 50; i++ {
        doc := fmt.Sprintf(`{
            "@id": "doc%d",
            "author": "test-author",
            "version": 1,
            "timestamp": "2023-01-01T00:00:00Z",
            "statements": [{
                "vulnerability": {"name": "CVE-2023-%04d"},
                "products": [{"component": {"@id": "product%d"}}],
                "status": "not_affected",
                "justification": "component_not_present"
            }]
        }`, i, 1000+i, i)
        documents[i] = doc
    }
    
    opts := &api.StreamMergeOptions{
        MergeOptions: api.MergeOptions{
            Documents: documents,
            OutputID:  "streamed-merge-test",
        },
        ChunkSize: 10,
    }
    
    resultCh, err := client.StreamMerge(context.Background(), opts)
    if err != nil {
        t.Fatalf("Failed to start stream merge: %v", err)
    }
    
    var progressUpdates []float64
    var finalDocument *api.VEXDocument
    var errorEncountered error
    
    for result := range resultCh {
        if result.Error != nil {
            errorEncountered = result.Error
            break
        }
        
        if result.Document != nil {
            finalDocument = result.Document
        } else {
            progressUpdates = append(progressUpdates, result.Progress)
        }
    }
    
    if errorEncountered != nil {
        t.Fatalf("Error during streaming merge: %v", errorEncountered)
    }
    
    if finalDocument == nil {
        t.Fatal("Expected final document but got none")
    }
    
    if len(finalDocument.Statements) != 50 {
        t.Errorf("Expected 50 statements, got %d", len(finalDocument.Statements))
    }
    
    if len(progressUpdates) < 3 {
        t.Errorf("Expected multiple progress updates, got %d", len(progressUpdates))
    }
    
    // Validate progress increases
    for i := 1; i < len(progressUpdates); i++ {
        if progressUpdates[i] < progressUpdates[i-1] {
            t.Errorf("Progress should increase, got %f after %f", progressUpdates[i], progressUpdates[i-1])
        }
    }
}

func setupTestClient(t *testing.T) *Client {
    config := &api.VEXConfig{
        DefaultAuthor: "test-default-author",
        ValidateOn:    true,
    }
    
    logger := helpers.NewTestLogger(t)
    return NewClient(config, logger)
}

func BenchmarkClient_CreateStatement(b *testing.B) {
    client := setupBenchmarkClient(b)
    
    opts := &api.CreateOptions{
        Product:       "nginx:1.20",
        Vulnerability: "CVE-2023-1234", 
        Status:        "not_affected",
        Justification: "component_not_present",
    }
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        _, err := client.CreateStatement(context.Background(), opts)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func setupBenchmarkClient(b *testing.B) *Client {
    config := &api.VEXConfig{
        DefaultAuthor: "bench-author",
        ValidateOn:    false, // Disable validation for benchmarks
    }
    
    logger := helpers.NewNullLogger()
    return NewClient(config, logger)
}
```

**Code Example - MCP Protocol Tests:**
```go
// internal/mcp/protocol_test.go
package mcp

import (
    "context"
    "encoding/json"
    "testing"
    
    "github.com/rosstaco/vexdoc-mcp-go/pkg/api"
    "github.com/rosstaco/vexdoc-mcp-go/test/helpers"
)

func TestProtocolHandler_HandleInitialize(t *testing.T) {
    handler := NewProtocolHandler(helpers.NewTestLogger(t))
    
    tests := []struct {
        name     string
        request  *api.Request
        wantErr  bool
        validate func(*testing.T, *api.Response)
    }{
        {
            name: "valid_initialize",
            request: &api.Request{
                JSONRPC: "2.0",
                ID:      1,
                Method:  "initialize",
                Params: map[string]interface{}{
                    "protocolVersion": "2024-11-05",
                    "clientInfo": map[string]interface{}{
                        "name":    "test-client",
                        "version": "1.0.0",
                    },
                },
            },
            wantErr: false,
            validate: func(t *testing.T, resp *api.Response) {
                if resp.Error != nil {
                    t.Errorf("Unexpected error: %v", resp.Error)
                }
                
                result, ok := resp.Result.(map[string]interface{})
                if !ok {
                    t.Fatal("Result should be a map")
                }
                
                if protocol, ok := result["protocolVersion"].(string); !ok || protocol != "2024-11-05" {
                    t.Errorf("Expected protocolVersion '2024-11-05', got %v", result["protocolVersion"])
                }
                
                if _, ok := result["serverInfo"]; !ok {
                    t.Error("Expected serverInfo in response")
                }
                
                if _, ok := result["capabilities"]; !ok {
                    t.Error("Expected capabilities in response")
                }
            },
        },
        {
            name: "unsupported_protocol_version",
            request: &api.Request{
                JSONRPC: "2.0",
                ID:      2,
                Method:  "initialize",
                Params: map[string]interface{}{
                    "protocolVersion": "unsupported-version",
                },
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resp := handler.HandleRequest(context.Background(), tt.request)
            
            if tt.wantErr {
                if resp.Error == nil {
                    t.Error("Expected error but got none")
                }
                return
            }
            
            if resp.Error != nil {
                t.Errorf("Unexpected error: %v", resp.Error)
                return
            }
            
            if tt.validate != nil {
                tt.validate(t, resp)
            }
        })
    }
}

func TestProtocolHandler_HandleToolsList(t *testing.T) {
    handler := NewProtocolHandler(helpers.NewTestLogger(t))
    
    // Register a test tool
    testTool := &helpers.MockTool{
        NameValue:        "test_tool",
        DescriptionValue: "A test tool",
        SchemaValue:      &api.JSONSchema{Type: "object"},
    }
    
    err := handler.RegisterTool(testTool)
    if err != nil {
        t.Fatalf("Failed to register test tool: %v", err)
    }
    
    request := &api.Request{
        JSONRPC: "2.0",
        ID:      1,
        Method:  "tools/list",
        Params:  nil,
    }
    
    resp := handler.HandleRequest(context.Background(), request)
    
    if resp.Error != nil {
        t.Fatalf("Unexpected error: %v", resp.Error)
    }
    
    result, ok := resp.Result.(map[string]interface{})
    if !ok {
        t.Fatal("Result should be a map")
    }
    
    tools, ok := result["tools"].([]map[string]interface{})
    if !ok {
        t.Fatal("Tools should be an array")
    }
    
    if len(tools) != 1 {
        t.Errorf("Expected 1 tool, got %d", len(tools))
    }
    
    if tools[0]["name"] != "test_tool" {
        t.Errorf("Expected tool name 'test_tool', got %v", tools[0]["name"])
    }
}

func TestProtocolHandler_HandleToolsCall(t *testing.T) {
    handler := NewProtocolHandler(helpers.NewTestLogger(t))
    
    // Register a test tool
    testTool := &helpers.MockTool{
        NameValue:        "test_tool",
        DescriptionValue: "A test tool",
        SchemaValue:      &api.JSONSchema{Type: "object"},
        ExecuteFunc: func(ctx context.Context, args map[string]interface{}) (*api.ToolResponse, error) {
            return &api.ToolResponse{
                Content: []api.Content{
                    {Type: "text", Text: "Test execution successful"},
                },
            }, nil
        },
    }
    
    err := handler.RegisterTool(testTool)
    if err != nil {
        t.Fatalf("Failed to register test tool: %v", err)
    }
    
    request := &api.Request{
        JSONRPC: "2.0",
        ID:      1,
        Method:  "tools/call",
        Params: map[string]interface{}{
            "name":      "test_tool",
            "arguments": map[string]interface{}{"test": "value"},
        },
    }
    
    resp := handler.HandleRequest(context.Background(), request)
    
    if resp.Error != nil {
        t.Fatalf("Unexpected error: %v", resp.Error)
    }
    
    toolResp, ok := resp.Result.(*api.ToolResponse)
    if !ok {
        t.Fatal("Result should be a ToolResponse")
    }
    
    if len(toolResp.Content) != 1 {
        t.Errorf("Expected 1 content item, got %d", len(toolResp.Content))
    }
    
    if toolResp.Content[0].Text != "Test execution successful" {
        t.Errorf("Unexpected content: %s", toolResp.Content[0].Text)
    }
}
```

**Deliverable**: Comprehensive unit test suite with >90% coverage

---

### Task 7.2: Integration Testing (3 points)

**Goal**: End-to-end integration tests validating complete system behavior

**Checklist:**
- [ ] Test complete MCP server startup and tool registration
- [ ] Test all transport modes (stdio, HTTP, streaming)
- [ ] Test real MCP client integration
- [ ] Test error handling across system boundaries
- [ ] Test concurrent operations and thread safety

**Code Example - Integration Tests:**
```go
// test/integration/server_test.go
package integration

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "os/exec"
    "strings"
    "testing"
    "time"
    
    "github.com/rosstaco/vexdoc-mcp-go/pkg/api"
)

func TestServer_StdioTransport(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Start server process
    cmd := exec.Command("go", "run", "../../cmd/server/main.go", "stdio")
    
    stdin, err := cmd.StdinPipe()
    if err != nil {
        t.Fatalf("Failed to get stdin pipe: %v", err)
    }
    
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        t.Fatalf("Failed to get stdout pipe: %v", err)
    }
    
    if err := cmd.Start(); err != nil {
        t.Fatalf("Failed to start server: %v", err)
    }
    
    defer func() {
        stdin.Close()
        cmd.Process.Kill()
        cmd.Wait()
    }()
    
    // Test initialize
    initRequest := &api.Request{
        JSONRPC: "2.0",
        ID:      1,
        Method:  "initialize",
        Params: map[string]interface{}{
            "protocolVersion": "2024-11-05",
            "clientInfo": map[string]interface{}{
                "name":    "integration-test",
                "version": "1.0.0",
            },
        },
    }
    
    response := sendRequest(t, stdin, stdout, initRequest)
    validateInitializeResponse(t, response)
    
    // Test tools/list
    listRequest := &api.Request{
        JSONRPC: "2.0",
        ID:      2,
        Method:  "tools/list",
        Params:  nil,
    }
    
    response = sendRequest(t, stdin, stdout, listRequest)
    validateToolsListResponse(t, response)
    
    // Test create_vex_statement tool
    createRequest := &api.Request{
        JSONRPC: "2.0",
        ID:      3,
        Method:  "tools/call",
        Params: map[string]interface{}{
            "name": "create_vex_statement",
            "arguments": map[string]interface{}{
                "product":       "nginx:1.20",
                "vulnerability": "CVE-2023-1234",
                "status":        "not_affected",
                "justification": "component_not_present",
            },
        },
    }
    
    response = sendRequest(t, stdin, stdout, createRequest)
    validateCreateVEXResponse(t, response)
}

func TestServer_HTTPTransport(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Start server in HTTP mode
    cmd := exec.Command("go", "run", "../../cmd/server/main.go", "http", "8080")
    
    if err := cmd.Start(); err != nil {
        t.Fatalf("Failed to start HTTP server: %v", err)
    }
    
    defer func() {
        cmd.Process.Kill()
        cmd.Wait()
    }()
    
    // Wait for server to start
    time.Sleep(2 * time.Second)
    
    baseURL := "http://localhost:8080"
    
    // Test initialize via HTTP
    initRequest := &api.Request{
        JSONRPC: "2.0",
        ID:      1,
        Method:  "initialize",
        Params: map[string]interface{}{
            "protocolVersion": "2024-11-05",
        },
    }
    
    response := sendHTTPRequest(t, baseURL+"/mcp", initRequest)
    validateInitializeResponse(t, response)
    
    // Test concurrent requests
    t.Run("concurrent_requests", func(t *testing.T) {
        const numRequests = 10
        responses := make(chan *api.Response, numRequests)
        
        for i := 0; i < numRequests; i++ {
            go func(id int) {
                req := &api.Request{
                    JSONRPC: "2.0",
                    ID:      id,
                    Method:  "tools/list",
                    Params:  nil,
                }
                resp := sendHTTPRequest(t, baseURL+"/mcp", req)
                responses <- resp
            }(i + 10)
        }
        
        // Collect all responses
        for i := 0; i < numRequests; i++ {
            select {
            case resp := <-responses:
                if resp.Error != nil {
                    t.Errorf("Request %v failed: %v", resp.ID, resp.Error)
                }
            case <-time.After(5 * time.Second):
                t.Fatal("Timeout waiting for concurrent responses")
            }
        }
    })
}

func TestServer_StreamingTransport(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Start server in streaming mode
    cmd := exec.Command("go", "run", "../../cmd/server/main.go", "streaming", "8081")
    
    if err := cmd.Start(); err != nil {
        t.Fatalf("Failed to start streaming server: %v", err)
    }
    
    defer func() {
        cmd.Process.Kill()
        cmd.Wait()
    }()
    
    time.Sleep(2 * time.Second)
    
    baseURL := "http://localhost:8081"
    
    // Test Server-Sent Events
    resp, err := http.Get(baseURL + "/mcp/sse?stream_id=test123")
    if err != nil {
        t.Fatalf("Failed to connect to SSE endpoint: %v", err)
    }
    defer resp.Body.Close()
    
    if resp.Header.Get("Content-Type") != "text/event-stream" {
        t.Error("Expected text/event-stream content type")
    }
    
    scanner := bufio.NewScanner(resp.Body)
    
    // Read connection event
    if scanner.Scan() {
        line := scanner.Text()
        if !strings.Contains(line, "connected") {
            t.Errorf("Expected connection event, got: %s", line)
        }
    }
    
    // Test streaming merge operation
    // This would require a more complex setup with document generation
}

func TestCompatibility_NodeJSOutput(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping compatibility test in short mode")
    }
    
    // Test that Go implementation produces identical output to Node.js
    testCases := []struct {
        name    string
        request *api.Request
        golden  string // Expected output file
    }{
        {
            name: "create_vex_statement_basic",
            request: &api.Request{
                JSONRPC: "2.0",
                ID:      1,
                Method:  "tools/call",
                Params: map[string]interface{}{
                    "name": "create_vex_statement",
                    "arguments": map[string]interface{}{
                        "product":       "nginx:1.20",
                        "vulnerability": "CVE-2023-1234",
                        "status":        "not_affected",
                        "justification": "component_not_present",
                    },
                },
            },
            golden: "testdata/create_basic.golden.json",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Get response from Go implementation
            goResponse := getGoResponse(t, tc.request)
            
            // Load expected output (could be from Node.js reference)
            expectedBytes, err := os.ReadFile(tc.golden)
            if err != nil {
                t.Fatalf("Failed to read golden file: %v", err)
            }
            
            var expectedResponse api.Response
            if err := json.Unmarshal(expectedBytes, &expectedResponse); err != nil {
                t.Fatalf("Failed to parse golden file: %v", err)
            }
            
            // Compare responses (excluding timestamps and generated IDs)
            compareResponses(t, goResponse, &expectedResponse)
        })
    }
}

// Helper functions
func sendRequest(t *testing.T, stdin *os.File, stdout *os.File, request *api.Request) *api.Response {
    // Send request
    requestBytes, err := json.Marshal(request)
    if err != nil {
        t.Fatalf("Failed to marshal request: %v", err)
    }
    
    _, err = fmt.Fprintf(stdin, "%s\n", string(requestBytes))
    if err != nil {
        t.Fatalf("Failed to send request: %v", err)
    }
    
    // Read response
    scanner := bufio.NewScanner(stdout)
    if !scanner.Scan() {
        t.Fatal("Failed to read response")
    }
    
    responseBytes := scanner.Bytes()
    var response api.Response
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        t.Fatalf("Failed to parse response: %v", err)
    }
    
    return &response
}

func sendHTTPRequest(t *testing.T, url string, request *api.Request) *api.Response {
    requestBytes, err := json.Marshal(request)
    if err != nil {
        t.Fatalf("Failed to marshal request: %v", err)
    }
    
    resp, err := http.Post(url, "application/json", strings.NewReader(string(requestBytes)))
    if err != nil {
        t.Fatalf("Failed to send HTTP request: %v", err)
    }
    defer resp.Body.Close()
    
    var response api.Response
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        t.Fatalf("Failed to parse HTTP response: %v", err)
    }
    
    return &response
}

func validateInitializeResponse(t *testing.T, response *api.Response) {
    if response.Error != nil {
        t.Fatalf("Initialize failed: %v", response.Error)
    }
    
    result, ok := response.Result.(map[string]interface{})
    if !ok {
        t.Fatal("Initialize result should be a map")
    }
    
    if protocol, ok := result["protocolVersion"].(string); !ok || protocol != "2024-11-05" {
        t.Errorf("Expected protocol version 2024-11-05, got %v", result["protocolVersion"])
    }
}

func validateToolsListResponse(t *testing.T, response *api.Response) {
    if response.Error != nil {
        t.Fatalf("Tools list failed: %v", response.Error)
    }
    
    result, ok := response.Result.(map[string]interface{})
    if !ok {
        t.Fatal("Tools list result should be a map")
    }
    
    tools, ok := result["tools"].([]interface{})
    if !ok {
        t.Fatal("Tools should be an array")
    }
    
    expectedTools := []string{"create_vex_statement", "merge_vex_documents"}
    if len(tools) != len(expectedTools) {
        t.Errorf("Expected %d tools, got %d", len(expectedTools), len(tools))
    }
    
    toolNames := make(map[string]bool)
    for _, tool := range tools {
        toolMap, ok := tool.(map[string]interface{})
        if !ok {
            t.Fatal("Tool should be a map")
        }
        
        name, ok := toolMap["name"].(string)
        if !ok {
            t.Fatal("Tool name should be a string")
        }
        
        toolNames[name] = true
    }
    
    for _, expectedTool := range expectedTools {
        if !toolNames[expectedTool] {
            t.Errorf("Expected tool %s not found", expectedTool)
        }
    }
}

func validateCreateVEXResponse(t *testing.T, response *api.Response) {
    if response.Error != nil {
        t.Fatalf("Create VEX failed: %v", response.Error)
    }
    
    toolResp, ok := response.Result.(*api.ToolResponse)
    if !ok {
        t.Fatal("Result should be a ToolResponse")
    }
    
    if len(toolResp.Content) < 1 {
        t.Fatal("Expected at least 1 content item")
    }
    
    if !strings.Contains(toolResp.Content[0].Text, "✅") {
        t.Error("Expected success indicator in response")
    }
}

func getGoResponse(t *testing.T, request *api.Request) *api.Response {
    // Implementation would start Go server and get response
    // Similar to sendHTTPRequest but with proper setup
    return nil
}

func compareResponses(t *testing.T, actual, expected *api.Response) {
    // Compare responses excluding dynamic fields like timestamps and IDs
    // Implementation would do deep comparison with exclusions
}
```

**Deliverable**: Complete integration test suite in `test/integration/`

---

### Task 7.3: Compatibility and Performance Validation (2 points)

**Goal**: Validate compatibility with Node.js output and performance targets

**Checklist:**
- [ ] Create output comparison tests against Node.js implementation
- [ ] Validate performance improvements meet 50%+ target
- [ ] Test memory usage under various load conditions
- [ ] Validate streaming operations with large datasets
- [ ] Create automated compatibility checks

**Code Example - Compatibility Tests:**
```go
// test/compatibility/nodejs_compat_test.go
package compatibility

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    "time"
    
    "github.com/rosstaco/vexdoc-mcp-go/pkg/api"
)

func TestNodeJSCompatibility(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping compatibility tests in short mode")
    }
    
    goldenDir := "testdata/nodejs_golden"
    
    testCases := []struct {
        name       string
        goldenFile string
        setupFunc  func() *api.Request
    }{
        {
            name:       "create_vex_not_affected",
            goldenFile: "create_not_affected.json",
            setupFunc: func() *api.Request {
                return &api.Request{
                    JSONRPC: "2.0",
                    ID:      1,
                    Method:  "tools/call",
                    Params: map[string]interface{}{
                        "name": "create_vex_statement",
                        "arguments": map[string]interface{}{
                            "product":       "nginx:1.20",
                            "vulnerability": "CVE-2023-1234",
                            "status":        "not_affected",
                            "justification": "component_not_present",
                        },
                    },
                }
            },
        },
        {
            name:       "create_vex_fixed",
            goldenFile: "create_fixed.json",
            setupFunc: func() *api.Request {
                return &api.Request{
                    JSONRPC: "2.0",
                    ID:      2,
                    Method:  "tools/call",
                    Params: map[string]interface{}{
                        "name": "create_vex_statement",
                        "arguments": map[string]interface{}{
                            "product":       "pkg:golang/github.com/gin-gonic/gin@v1.9.1",
                            "vulnerability": "CVE-2023-5678",
                            "status":        "fixed",
                        },
                    },
                }
            },
        },
        {
            name:       "merge_two_documents",
            goldenFile: "merge_simple.json",
            setupFunc:  setupMergeTest,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Load Node.js golden output
            goldenPath := filepath.Join(goldenDir, tc.goldenFile)
            goldenBytes, err := os.ReadFile(goldenPath)
            if err != nil {
                t.Fatalf("Failed to read golden file: %v", err)
            }
            
            var goldenResp NodeJSResponse
            if err := json.Unmarshal(goldenBytes, &goldenResp); err != nil {
                t.Fatalf("Failed to parse golden response: %v", err)
            }
            
            // Get Go implementation response
            request := tc.setupFunc()
            goResp := executeGoRequest(t, request)
            
            // Compare responses
            compareWithNodeJS(t, goResp, &goldenResp)
        })
    }
}

type NodeJSResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Result  struct {
        Content []struct {
            Type string `json:"type"`
            Text string `json:"text"`
        } `json:"content"`
        IsError bool `json:"isError,omitempty"`
    } `json:"result,omitempty"`
    Error *struct {
        Code    int    `json:"code"`
        Message string `json:"message"`
    } `json:"error,omitempty"`
}

func compareWithNodeJS(t *testing.T, goResp *api.Response, nodeResp *NodeJSResponse) {
    // Compare JSON-RPC structure
    if goResp.JSONRPC != nodeResp.JSONRPC {
        t.Errorf("JSONRPC mismatch: Go=%s, Node=%s", goResp.JSONRPC, nodeResp.JSONRPC)
    }
    
    // Compare error states
    if (goResp.Error == nil) != (nodeResp.Error == nil) {
        t.Errorf("Error state mismatch: Go has error=%t, Node has error=%t", 
            goResp.Error != nil, nodeResp.Error != nil)
    }
    
    if goResp.Error != nil && nodeResp.Error != nil {
        if goResp.Error.Message != nodeResp.Error.Message {
            t.Errorf("Error message mismatch: Go=%s, Node=%s", 
                goResp.Error.Message, nodeResp.Error.Message)
        }
        return
    }
    
    // Compare successful responses
    if goResp.Result == nil {
        t.Fatal("Go response missing result")
    }
    
    toolResp, ok := goResp.Result.(*api.ToolResponse)
    if !ok {
        t.Fatal("Go result should be ToolResponse")
    }
    
    if len(toolResp.Content) != len(nodeResp.Result.Content) {
        t.Errorf("Content length mismatch: Go=%d, Node=%d", 
            len(toolResp.Content), len(nodeResp.Result.Content))
    }
    
    // Compare content structure (excluding dynamic values)
    for i, goContent := range toolResp.Content {
        if i >= len(nodeResp.Result.Content) {
            break
        }
        
        nodeContent := nodeResp.Result.Content[i]
        
        if goContent.Type != nodeContent.Type {
            t.Errorf("Content[%d] type mismatch: Go=%s, Node=%s", 
                i, goContent.Type, nodeContent.Type)
        }
        
        // For VEX documents, compare structure not exact strings
        if goContent.Type == "text" && nodeContent.Type == "text" {
            compareVEXContent(t, goContent.Text, nodeContent.Text)
        }
    }
}

func compareVEXContent(t *testing.T, goText, nodeText string) {
    // Parse JSON if it looks like VEX document
    if strings.Contains(goText, `"@id"`) && strings.Contains(nodeText, `"@id"`) {
        var goDoc, nodeDoc map[string]interface{}
        
        if err := json.Unmarshal([]byte(goText), &goDoc); err != nil {
            t.Logf("Go text not JSON: %s", goText)
            return
        }
        
        if err := json.Unmarshal([]byte(nodeText), &nodeDoc); err != nil {
            t.Logf("Node text not JSON: %s", nodeText)
            return
        }
        
        // Compare structure excluding dynamic fields
        compareVEXDocuments(t, goDoc, nodeDoc)
    }
}

func compareVEXDocuments(t *testing.T, goDoc, nodeDoc map[string]interface{}) {
    // Compare static fields
    staticFields := []string{"author", "version"}
    
    for _, field := range staticFields {
        if goDoc[field] != nodeDoc[field] {
            t.Errorf("Field %s mismatch: Go=%v, Node=%v", field, goDoc[field], nodeDoc[field])
        }
    }
    
    // Compare statements structure
    goStatements, ok1 := goDoc["statements"].([]interface{})
    nodeStatements, ok2 := nodeDoc["statements"].([]interface{})
    
    if ok1 && ok2 {
        if len(goStatements) != len(nodeStatements) {
            t.Errorf("Statements count mismatch: Go=%d, Node=%d", 
                len(goStatements), len(nodeStatements))
        }
        
        // Compare first statement structure
        if len(goStatements) > 0 && len(nodeStatements) > 0 {
            compareStatements(t, goStatements[0], nodeStatements[0])
        }
    }
}

func compareStatements(t *testing.T, goStmt, nodeStmt interface{}) {
    goMap, ok1 := goStmt.(map[string]interface{})
    nodeMap, ok2 := nodeStmt.(map[string]interface{})
    
    if !ok1 || !ok2 {
        return
    }
    
    // Compare status
    if goMap["status"] != nodeMap["status"] {
        t.Errorf("Statement status mismatch: Go=%v, Node=%v", 
            goMap["status"], nodeMap["status"])
    }
    
    // Compare justification if present
    if goJust, ok := goMap["justification"]; ok {
        if nodeJust, ok := nodeMap["justification"]; ok {
            if goJust != nodeJust {
                t.Errorf("Statement justification mismatch: Go=%v, Node=%v", 
                    goJust, nodeJust)
            }
        }
    }
}

// Performance comparison tests
func TestPerformanceComparison(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance tests in short mode")
    }
    
    testCases := []struct {
        name            string
        operation       func() error
        targetTime      time.Duration
        maxMemoryMB     int
    }{
        {
            name: "create_vex_statement",
            operation: func() error {
                return executeCreateVEXStatement()
            },
            targetTime:  10 * time.Millisecond,
            maxMemoryMB: 5,
        },
        {
            name: "merge_100_documents",
            operation: func() error {
                return executeMerge100Documents()
            },
            targetTime:  1 * time.Second,
            maxMemoryMB: 50,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Measure execution time
            start := time.Now()
            err := tc.operation()
            duration := time.Since(start)
            
            if err != nil {
                t.Fatalf("Operation failed: %v", err)
            }
            
            if duration > tc.targetTime {
                t.Errorf("Operation too slow: %v > %v", duration, tc.targetTime)
            }
            
            t.Logf("Operation completed in %v (target: %v)", duration, tc.targetTime)
        })
    }
}
```

**Deliverable**: Compatibility validation suite in `test/compatibility/`

---

## Phase 7 Deliverables

### 1. Unit Test Suite (`internal/*/test.go`)
- [ ] >90% test coverage across all packages
- [ ] Unit tests for VEX client, MCP protocol, tools
- [ ] Error handling and edge case coverage
- [ ] Performance benchmarks for critical paths

### 2. Integration Test Suite (`test/integration/`)
- [ ] End-to-end server functionality tests
- [ ] All transport mode validation
- [ ] Concurrent operation testing
- [ ] Real client integration tests

### 3. Compatibility Validation (`test/compatibility/`)
- [ ] Output format comparison with Node.js
- [ ] Golden file testing for consistent output
- [ ] API compatibility verification
- [ ] Performance comparison validation

### 4. Test Infrastructure (`test/helpers/`)
- [ ] Mock implementations for testing
- [ ] Test utilities and helpers
- [ ] Test data generators
- [ ] CI/CD integration scripts

### 5. Documentation
- [ ] Testing guide and best practices
- [ ] Performance benchmark results
- [ ] Compatibility validation report
- [ ] Test coverage reports

## Success Criteria
- [ ] >90% test coverage across all packages
- [ ] All integration tests pass consistently
- [ ] Output format exactly matches Node.js implementation
- [ ] Performance targets met (50%+ improvement)
- [ ] Memory usage within acceptable limits
- [ ] Zero data races or concurrency issues
- [ ] All error scenarios properly handled

## Test Execution Strategy
```bash
# Unit tests
make test-unit

# Integration tests  
make test-integration

# Compatibility tests
make test-compatibility

# Performance tests
make test-performance

# Full test suite
make test-all

# Coverage report
make test-coverage
```

## Quality Gates
- **Unit Tests**: 90%+ coverage, all tests pass
- **Integration Tests**: All transport modes work, concurrent operations stable
- **Compatibility**: 100% output format compatibility with Node.js
- **Performance**: 50%+ improvement in execution time
- **Memory**: Stable usage under load, no memory leaks

## Dependencies
- **Input**: Phase 6 performance-optimized implementation
- **Output**: Production-ready system with comprehensive test validation

## Risks & Mitigation
- **Risk**: Test complexity slows development
  - **Mitigation**: Focus on critical paths first, expand coverage iteratively
- **Risk**: Node.js compatibility edge cases
  - **Mitigation**: Extensive golden file testing with real Node.js output
- **Risk**: Performance tests are environment-dependent
  - **Mitigation**: Relative performance measurements, multiple test runs

## Time Estimate
**8 Story Points** ≈ 3-4 days of focused development

---
**Next**: [Phase 8: Deployment & Migration](./phase8-deployment.md)
