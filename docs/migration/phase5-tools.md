# Phase 5: Tool Implementation
**Story Points**: 8 | **Prerequisites**: [Phase 4](./phase4-vex-integration.md) | **Next**: [Phase 6](./phase6-streaming.md)

## Overview
Implement the MCP tools (`create_vex_statement` and `merge_vex_documents`) using the native VEX integration, ensuring feature parity with the Node.js implementation.

## Objectives
- [ ] Implement `create_vex_statement` tool with native VEX integration
- [ ] Implement `merge_vex_documents` tool with streaming support
- [ ] Ensure exact API compatibility with Node.js version
- [ ] Add comprehensive input validation and error handling
- [ ] Create tool registry and registration system

## Tasks

### Task 5.1: Create VEX Statement Tool (3 points)

**Goal**: Implement the `create_vex_statement` tool with native VEX library

**Checklist:**
- [ ] Create tool structure implementing the Tool interface
- [ ] Add comprehensive input validation
- [ ] Integrate with native VEX client
- [ ] Ensure output format matches Node.js version
- [ ] Add error handling and logging

**Code Example - Create VEX Statement Tool:**
```go
// internal/tools/vex_create.go
package tools

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/rosstaco/vexdoc-mcp/pkg/api"
    "github.com/rosstaco/vexdoc-mcp/internal/vex"
    "github.com/rosstaco/vexdoc-mcp/internal/logging"
    "github.com/rosstaco/vexdoc-mcp/internal/errors"
)

type CreateVEXStatementTool struct {
    vexClient api.VEXClient
    logger    logging.Logger
}

func NewCreateVEXStatementTool(vexClient api.VEXClient, logger logging.Logger) *CreateVEXStatementTool {
    return &CreateVEXStatementTool{
        vexClient: vexClient,
        logger:    logger,
    }
}

func (t *CreateVEXStatementTool) Name() string {
    return "create_vex_statement"
}

func (t *CreateVEXStatementTool) Description() string {
    return "Create a VEX (Vulnerability Exploitability eXchange) statement for a product and vulnerability"
}

func (t *CreateVEXStatementTool) InputSchema() *api.JSONSchema {
    return &api.JSONSchema{
        Type: "object",
        Properties: map[string]interface{}{
            "product": map[string]interface{}{
                "type":        "string",
                "description": "Product identifier (e.g., package name, container image)",
                "examples":    []string{"nginx:1.20", "pkg:golang/github.com/gin-gonic/gin@v1.9.1"},
            },
            "vulnerability": map[string]interface{}{
                "type":        "string", 
                "description": "Vulnerability identifier (e.g., CVE-2023-1234)",
                "pattern":     "^CVE-\\d{4}-\\d{4,}$",
                "examples":    []string{"CVE-2023-1234", "CVE-2024-5678"},
            },
            "status": map[string]interface{}{
                "type":        "string",
                "description": "VEX status for the vulnerability",
                "enum":        []string{"not_affected", "affected", "fixed", "under_investigation"},
            },
            "justification": map[string]interface{}{
                "type":        "string",
                "description": "Justification for the status (required for not_affected)",
                "examples":    []string{"component_not_present", "vulnerable_code_not_present", "vulnerable_code_not_in_execute_path"},
            },
            "impact": map[string]interface{}{
                "type":        "string",
                "description": "Impact assessment (optional)",
                "examples":    []string{"high", "medium", "low"},
            },
            "author": map[string]interface{}{
                "type":        "string",
                "description": "Author of the VEX statement (optional, uses default if not provided)",
                "examples":    []string{"security-team@company.com", "John Doe <john@example.com>"},
            },
        },
        Required: []string{"product", "vulnerability", "status"},
    }
}

func (t *CreateVEXStatementTool) Execute(ctx context.Context, args map[string]interface{}) (*api.ToolResponse, error) {
    logger := logging.FromContext(ctx, t.logger)
    logger.Info("Executing create_vex_statement tool")
    
    // Validate and parse arguments
    opts, err := t.parseCreateOptions(args)
    if err != nil {
        return nil, errors.NewValidationError("Invalid arguments", err)
    }
    
    // Additional validation
    if err := t.validateCreateOptions(opts); err != nil {
        return nil, errors.NewValidationError("Validation failed", err)
    }
    
    logger.Info("Creating VEX statement",
        "product", opts.Product,
        "vulnerability", opts.Vulnerability,
        "status", opts.Status,
    )
    
    // Create VEX statement using native client
    document, err := t.vexClient.CreateStatement(ctx, opts)
    if err != nil {
        return nil, errors.NewVEXError("Failed to create VEX statement", err)
    }
    
    // Format response
    content, err := t.formatResponse(document)
    if err != nil {
        return nil, errors.NewInternalError("Failed to format response", err)
    }
    
    logger.Info("VEX statement created successfully", "document_id", document.ID)
    
    return &api.ToolResponse{
        Content: []api.Content{
            {
                Type: "text",
                Text: fmt.Sprintf("✅ VEX statement created successfully for %s", opts.Product),
            },
            {
                Type: "text", 
                Text: content,
            },
        },
    }, nil
}

func (t *CreateVEXStatementTool) parseCreateOptions(args map[string]interface{}) (*api.CreateOptions, error) {
    opts := &api.CreateOptions{}
    
    // Required fields
    product, ok := args["product"].(string)
    if !ok || product == "" {
        return nil, fmt.Errorf("product is required and must be a non-empty string")
    }
    opts.Product = product
    
    vulnerability, ok := args["vulnerability"].(string)
    if !ok || vulnerability == "" {
        return nil, fmt.Errorf("vulnerability is required and must be a non-empty string")
    }
    opts.Vulnerability = vulnerability
    
    status, ok := args["status"].(string)
    if !ok || status == "" {
        return nil, fmt.Errorf("status is required and must be a non-empty string")
    }
    opts.Status = status
    
    // Optional fields
    if justification, ok := args["justification"].(string); ok {
        opts.Justification = justification
    }
    
    if impact, ok := args["impact"].(string); ok {
        opts.Impact = impact
    }
    
    if author, ok := args["author"].(string); ok {
        opts.Author = author
    }
    
    return opts, nil
}

func (t *CreateVEXStatementTool) validateCreateOptions(opts *api.CreateOptions) error {
    // Validate status
    validStatuses := map[string]bool{
        "not_affected":        true,
        "affected":            true,
        "fixed":               true,
        "under_investigation": true,
    }
    
    if !validStatuses[opts.Status] {
        return fmt.Errorf("invalid status: %s (must be one of: not_affected, affected, fixed, under_investigation)", opts.Status)
    }
    
    // Validate CVE format
    if !t.isValidCVE(opts.Vulnerability) {
        return fmt.Errorf("invalid vulnerability format: %s (expected CVE-YYYY-NNNN format)", opts.Vulnerability)
    }
    
    // Justification required for not_affected status
    if opts.Status == "not_affected" && opts.Justification == "" {
        return fmt.Errorf("justification is required when status is 'not_affected'")
    }
    
    // Validate impact if provided
    if opts.Impact != "" {
        validImpacts := map[string]bool{
            "high": true, "medium": true, "low": true,
        }
        if !validImpacts[opts.Impact] {
            return fmt.Errorf("invalid impact: %s (must be one of: high, medium, low)", opts.Impact)
        }
    }
    
    return nil
}

func (t *CreateVEXStatementTool) isValidCVE(cve string) bool {
    // Basic CVE format validation: CVE-YYYY-NNNN (at least 4 digits after year)
    if len(cve) < 13 { // CVE-2023-1234 = 13 chars minimum
        return false
    }
    
    if cve[:4] != "CVE-" {
        return false
    }
    
    // More sophisticated validation could be added here
    return true
}

func (t *CreateVEXStatementTool) formatResponse(document *api.VEXDocument) (string, error) {
    // Format as pretty JSON to match Node.js output
    jsonBytes, err := json.MarshalIndent(document, "", "  ")
    if err != nil {
        return "", fmt.Errorf("marshaling document: %w", err)
    }
    
    return string(jsonBytes), nil
}
```

**Deliverable**: Create VEX statement tool in `internal/tools/vex_create.go`

---

### Task 5.2: Merge VEX Documents Tool (3 points)

**Goal**: Implement the `merge_vex_documents` tool with streaming support

**Checklist:**
- [ ] Create merge tool implementing both Tool and StreamingTool interfaces
- [ ] Add document validation and parsing
- [ ] Integrate with streaming VEX merge capabilities
- [ ] Support multiple input formats (JSON strings, file paths)
- [ ] Add progress reporting for large merges

**Code Example - Merge VEX Documents Tool:**
```go
// internal/tools/vex_merge.go
package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    
    "github.com/rosstaco/vexdoc-mcp/pkg/api"
    "github.com/rosstaco/vexdoc-mcp/internal/vex"
    "github.com/rosstaco/vexdoc-mcp/internal/logging"
    "github.com/rosstaco/vexdoc-mcp/internal/errors"
)

type MergeVEXDocumentsTool struct {
    vexClient api.VEXClient
    logger    logging.Logger
}

func NewMergeVEXDocumentsTool(vexClient api.VEXClient, logger logging.Logger) *MergeVEXDocumentsTool {
    return &MergeVEXDocumentsTool{
        vexClient: vexClient,
        logger:    logger,
    }
}

func (t *MergeVEXDocumentsTool) Name() string {
    return "merge_vex_documents"
}

func (t *MergeVEXDocumentsTool) Description() string {
    return "Merge multiple VEX documents into a single document with conflict resolution"
}

func (t *MergeVEXDocumentsTool) InputSchema() *api.JSONSchema {
    return &api.JSONSchema{
        Type: "object",
        Properties: map[string]interface{}{
            "documents": map[string]interface{}{
                "type":        "array",
                "description": "Array of VEX documents to merge (as JSON strings)",
                "items": map[string]interface{}{
                    "type": "string",
                    "description": "VEX document as JSON string",
                },
                "minItems": 2,
            },
            "output_id": map[string]interface{}{
                "type":        "string",
                "description": "Optional ID for the merged document",
            },
            "author": map[string]interface{}{
                "type":        "string",
                "description": "Author for the merged document",
            },
            "streaming": map[string]interface{}{
                "type":        "boolean",
                "description": "Enable streaming mode for large document sets",
                "default":     false,
            },
            "chunk_size": map[string]interface{}{
                "type":        "integer",
                "description": "Number of documents to process in each chunk (streaming mode)",
                "default":     10,
                "minimum":     1,
            },
        },
        Required: []string{"documents"},
    }
}

func (t *MergeVEXDocumentsTool) Execute(ctx context.Context, args map[string]interface{}) (*api.ToolResponse, error) {
    logger := logging.FromContext(ctx, t.logger)
    logger.Info("Executing merge_vex_documents tool")
    
    // Check if streaming is requested
    if streaming, ok := args["streaming"].(bool); ok && streaming {
        return t.executeStreaming(ctx, args)
    }
    
    // Parse arguments
    opts, err := t.parseMergeOptions(args)
    if err != nil {
        return nil, errors.NewValidationError("Invalid arguments", err)
    }
    
    // Validate documents
    if err := t.validateDocuments(opts.Documents); err != nil {
        return nil, errors.NewValidationError("Document validation failed", err)
    }
    
    logger.Info("Merging VEX documents", "count", len(opts.Documents))
    
    // Perform merge
    merged, err := t.vexClient.MergeDocuments(ctx, opts)
    if err != nil {
        return nil, errors.NewVEXError("Failed to merge documents", err)
    }
    
    // Format response
    content, err := t.formatMergeResponse(merged, len(opts.Documents))
    if err != nil {
        return nil, errors.NewInternalError("Failed to format response", err)
    }
    
    logger.Info("VEX documents merged successfully",
        "input_count", len(opts.Documents),
        "output_statements", len(merged.Statements),
    )
    
    return &api.ToolResponse{
        Content: []api.Content{
            {
                Type: "text",
                Text: fmt.Sprintf("✅ Successfully merged %d VEX documents into 1 document with %d statements", 
                    len(opts.Documents), len(merged.Statements)),
            },
            {
                Type: "text",
                Text: content,
            },
        },
    }, nil
}

func (t *MergeVEXDocumentsTool) StreamExecute(ctx context.Context, args map[string]interface{}) (<-chan *api.ToolResponse, error) {
    logger := logging.FromContext(ctx, t.logger)
    logger.Info("Executing streaming merge_vex_documents tool")
    
    // Parse streaming options
    opts, err := t.parseStreamMergeOptions(args)
    if err != nil {
        return nil, errors.NewValidationError("Invalid streaming arguments", err)
    }
    
    // Start streaming merge
    mergeCh, err := t.vexClient.StreamMerge(ctx, opts)
    if err != nil {
        return nil, errors.NewVEXError("Failed to start streaming merge", err)
    }
    
    // Create response channel
    responseCh := make(chan *api.ToolResponse, 10)
    
    go func() {
        defer close(responseCh)
        t.handleStreamingMerge(ctx, mergeCh, responseCh, len(opts.Documents))
    }()
    
    return responseCh, nil
}

func (t *MergeVEXDocumentsTool) executeStreaming(ctx context.Context, args map[string]interface{}) (*api.ToolResponse, error) {
    // For non-streaming interface, collect all streaming results
    streamCh, err := t.StreamExecute(ctx, args)
    if err != nil {
        return nil, err
    }
    
    var lastResponse *api.ToolResponse
    var progressMessages []string
    
    for response := range streamCh {
        if response.IsError {
            return response, nil
        }
        
        // Collect progress messages
        if len(response.Content) > 0 && response.Content[0].Text != "" {
            if strings.Contains(response.Content[0].Text, "%") {
                progressMessages = append(progressMessages, response.Content[0].Text)
            }
        }
        
        lastResponse = response
    }
    
    // Combine progress messages with final result
    if lastResponse != nil && len(progressMessages) > 0 {
        progressSummary := fmt.Sprintf("Progress: %s", strings.Join(progressMessages, " → "))
        lastResponse.Content = append([]api.Content{
            {Type: "text", Text: progressSummary},
        }, lastResponse.Content...)
    }
    
    return lastResponse, nil
}

func (t *MergeVEXDocumentsTool) parseMergeOptions(args map[string]interface{}) (*api.MergeOptions, error) {
    opts := &api.MergeOptions{}
    
    // Parse documents array
    docsInterface, ok := args["documents"]
    if !ok {
        return nil, fmt.Errorf("documents field is required")
    }
    
    docsArray, ok := docsInterface.([]interface{})
    if !ok {
        return nil, fmt.Errorf("documents must be an array")
    }
    
    if len(docsArray) < 2 {
        return nil, fmt.Errorf("at least 2 documents are required for merge")
    }
    
    for i, doc := range docsArray {
        docStr, ok := doc.(string)
        if !ok {
            return nil, fmt.Errorf("document %d must be a string", i)
        }
        opts.Documents = append(opts.Documents, docStr)
    }
    
    // Optional fields
    if outputID, ok := args["output_id"].(string); ok {
        opts.OutputID = outputID
    }
    
    if author, ok := args["author"].(string); ok {
        opts.Author = author
    }
    
    return opts, nil
}

func (t *MergeVEXDocumentsTool) parseStreamMergeOptions(args map[string]interface{}) (*api.StreamMergeOptions, error) {
    baseOpts, err := t.parseMergeOptions(args)
    if err != nil {
        return nil, err
    }
    
    opts := &api.StreamMergeOptions{
        MergeOptions: *baseOpts,
        ChunkSize:    10, // Default
    }
    
    if chunkSize, ok := args["chunk_size"].(float64); ok {
        opts.ChunkSize = int(chunkSize)
    }
    
    if opts.ChunkSize < 1 {
        opts.ChunkSize = 1
    }
    
    return opts, nil
}

func (t *MergeVEXDocumentsTool) validateDocuments(documents []string) error {
    for i, docStr := range documents {
        // Try to parse as JSON to validate structure
        var doc map[string]interface{}
        if err := json.Unmarshal([]byte(docStr), &doc); err != nil {
            return fmt.Errorf("document %d is not valid JSON: %w", i, err)
        }
        
        // Basic VEX document validation
        if _, ok := doc["statements"]; !ok {
            return fmt.Errorf("document %d missing required 'statements' field", i)
        }
        
        if statements, ok := doc["statements"].([]interface{}); ok {
            if len(statements) == 0 {
                return fmt.Errorf("document %d has no statements", i)
            }
        }
    }
    
    return nil
}

func (t *MergeVEXDocumentsTool) handleStreamingMerge(ctx context.Context, mergeCh <-chan *api.MergeResult, responseCh chan<- *api.ToolResponse, inputCount int) {
    logger := logging.FromContext(ctx, t.logger)
    
    for result := range mergeCh {
        if result.Error != nil {
            responseCh <- &api.ToolResponse{
                Content: []api.Content{
                    {
                        Type: "text", 
                        Text: fmt.Sprintf("❌ Merge failed: %v", result.Error),
                    },
                },
                IsError: true,
            }
            return
        }
        
        if result.Document != nil {
            // Final result
            content, err := t.formatMergeResponse(result.Document, inputCount)
            if err != nil {
                logger.Error("Failed to format final response", "error", err)
                responseCh <- &api.ToolResponse{
                    Content: []api.Content{
                        {Type: "text", Text: fmt.Sprintf("❌ Failed to format response: %v", err)},
                    },
                    IsError: true,
                }
                return
            }
            
            responseCh <- &api.ToolResponse{
                Content: []api.Content{
                    {
                        Type: "text",
                        Text: fmt.Sprintf("✅ Successfully merged %d VEX documents into 1 document with %d statements",
                            inputCount, len(result.Document.Statements)),
                    },
                    {
                        Type: "text",
                        Text: content,
                    },
                },
            }
        } else {
            // Progress update
            progressPercent := int(result.Progress * 100)
            responseCh <- &api.ToolResponse{
                Content: []api.Content{
                    {
                        Type: "text",
                        Text: fmt.Sprintf("🔄 %s (%d%%)", result.Status, progressPercent),
                    },
                },
            }
        }
    }
}

func (t *MergeVEXDocumentsTool) formatMergeResponse(document *api.VEXDocument, inputCount int) (string, error) {
    // Create summary information
    summary := map[string]interface{}{
        "merge_summary": map[string]interface{}{
            "input_documents": inputCount,
            "output_statements": len(document.Statements),
            "document_id": document.ID,
            "merged_at": document.Timestamp,
        },
        "merged_document": document,
    }
    
    jsonBytes, err := json.MarshalIndent(summary, "", "  ")
    if err != nil {
        return "", fmt.Errorf("marshaling merge result: %w", err)
    }
    
    return string(jsonBytes), nil
}
```

**Deliverable**: Merge VEX documents tool in `internal/tools/vex_merge.go`

---

### Task 5.3: Tool Registry and Integration (2 points)

**Goal**: Create tool registry and integrate tools with MCP server

**Checklist:**
- [ ] Create tool registry for managing available tools
- [ ] Integrate tools with MCP server
- [ ] Add tool discovery and registration
- [ ] Update server startup to register tools
- [ ] Add tool configuration and initialization

**Code Example - Tool Registry:**
```go
// internal/tools/registry.go
package tools

import (
    "fmt"
    "sync"
    
    "github.com/rosstaco/vexdoc-mcp/pkg/api"
    "github.com/rosstaco/vexdoc-mcp/internal/logging"
)

type Registry struct {
    tools  map[string]api.Tool
    logger logging.Logger
    mu     sync.RWMutex
}

func NewRegistry(logger logging.Logger) *Registry {
    return &Registry{
        tools:  make(map[string]api.Tool),
        logger: logger,
    }
}

func (r *Registry) Register(tool api.Tool) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    name := tool.Name()
    if _, exists := r.tools[name]; exists {
        return fmt.Errorf("tool already registered: %s", name)
    }
    
    r.tools[name] = tool
    r.logger.Info("Tool registered", "tool_name", name)
    
    return nil
}

func (r *Registry) Get(name string) (api.Tool, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    tool, exists := r.tools[name]
    return tool, exists
}

func (r *Registry) List() []api.Tool {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    tools := make([]api.Tool, 0, len(r.tools))
    for _, tool := range r.tools {
        tools = append(tools, tool)
    }
    
    return tools
}

func (r *Registry) RegisterVEXTools(vexClient api.VEXClient) error {
    // Register create VEX statement tool
    createTool := NewCreateVEXStatementTool(vexClient, r.logger)
    if err := r.Register(createTool); err != nil {
        return fmt.Errorf("registering create tool: %w", err)
    }
    
    // Register merge VEX documents tool  
    mergeTool := NewMergeVEXDocumentsTool(vexClient, r.logger)
    if err := r.Register(mergeTool); err != nil {
        return fmt.Errorf("registering merge tool: %w", err)
    }
    
    r.logger.Info("All VEX tools registered successfully")
    return nil
}
```

**Code Example - Server Integration:**
```go
// cmd/server/main.go (updated)
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/rosstaco/vexdoc-mcp/internal/config"
    "github.com/rosstaco/vexdoc-mcp/internal/logging"
    "github.com/rosstaco/vexdoc-mcp/internal/mcp"
    "github.com/rosstaco/vexdoc-mcp/internal/vex"
    "github.com/rosstaco/vexdoc-mcp/internal/tools"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
    
    // Initialize logger
    logger, err := logging.New(cfg.Logging.Level, cfg.Logging.Format, os.Stderr)
    if err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }
    
    logger.Info("Starting VEX MCP Server",
        "name", cfg.Name,
        "version", cfg.Version,
        "transport", cfg.Transport.Type,
    )
    
    // Create VEX client
    vexClient := vex.NewClient(&cfg.VEX, logger)
    
    // Create tool registry and register VEX tools
    toolRegistry := tools.NewRegistry(logger)
    if err := toolRegistry.RegisterVEXTools(vexClient); err != nil {
        logger.Error("Failed to register VEX tools", "error", err)
        os.Exit(1)
    }
    
    // Create MCP server
    server, err := mcp.NewServer(cfg, logger)
    if err != nil {
        logger.Error("Failed to create server", "error", err)
        os.Exit(1)
    }
    
    // Register all tools with MCP server
    for _, tool := range toolRegistry.List() {
        if err := server.RegisterTool(tool); err != nil {
            logger.Error("Failed to register tool", "tool", tool.Name(), "error", err)
            os.Exit(1)
        }
    }
    
    // Start server
    if err := server.Start(); err != nil {
        logger.Error("Failed to start server", "error", err)
        os.Exit(1)
    }
    
    logger.Info("VEX MCP Server started successfully")
    
    // Wait for server to finish
    server.Wait()
    
    logger.Info("VEX MCP Server stopped")
}
```

**Deliverable**: Tool registry and server integration

---

## Phase 5 Deliverables

### 1. Create VEX Statement Tool (`internal/tools/vex_create.go`)
- [ ] Complete tool implementation with validation
- [ ] Native VEX client integration
- [ ] Input schema and error handling
- [ ] Output formatting matching Node.js version

### 2. Merge VEX Documents Tool (`internal/tools/vex_merge.go`) 
- [ ] Standard and streaming merge implementations
- [ ] Document validation and parsing
- [ ] Progress reporting for streaming operations
- [ ] Conflict resolution with status priorities

### 3. Tool Registry (`internal/tools/registry.go`)
- [ ] Tool registration and discovery system
- [ ] Thread-safe tool management
- [ ] Integration with MCP server
- [ ] VEX-specific tool initialization

### 4. Server Integration (`cmd/server/main.go`)
- [ ] Updated main function with tool registration
- [ ] VEX client initialization
- [ ] Complete startup sequence
- [ ] Error handling and logging

### 5. Integration Tests
- [ ] End-to-end tool execution tests
- [ ] Streaming operation validation
- [ ] Error handling scenarios
- [ ] Compatibility tests with Node.js output

## Success Criteria
- [ ] All tools execute successfully via MCP protocol
- [ ] Output format exactly matches Node.js version
- [ ] Streaming operations work for large document sets
- [ ] Error messages are helpful and actionable
- [ ] Performance improvement over Node.js subprocess approach

## API Compatibility Validation
```bash
# Test create tool
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"create_vex_statement","arguments":{"product":"nginx:1.20","vulnerability":"CVE-2023-1234","status":"not_affected","justification":"component_not_present"}}}' | go run cmd/server/main.go

# Test merge tool  
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"merge_vex_documents","arguments":{"documents":["{}","{}"],"streaming":false}}}' | go run cmd/server/main.go
```

## Dependencies
- **Input**: Phase 4 native VEX integration
- **Output**: Feature-complete MCP tools ready for performance optimization

## Risks & Mitigation
- **Risk**: Tool output format differences from Node.js
  - **Mitigation**: Extensive comparison testing with existing outputs
- **Risk**: Streaming complexity in tool interface
  - **Mitigation**: Clear separation between streaming and non-streaming modes
- **Risk**: Error handling inconsistencies
  - **Mitigation**: Standardized error response formats

## Time Estimate
**8 Story Points** ≈ 3-4 days of focused development

---
**Next**: [Phase 6: Streaming & Performance](./phase6-streaming.md)
