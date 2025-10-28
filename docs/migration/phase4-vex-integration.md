# Phase 4: VEX Native Integration
**Story Points**: 5 | **Prerequisites**: [Phase 3](./phase3-mcp-core.md) | **Next**: [Phase 5](./phase5-tools.md)

## Overview
Implement native Go integration with VEX libraries, replacing subprocess calls with direct API usage for better performance and streaming capabilities.

## Objectives
- [ ] Replace subprocess vexctl calls with native Go VEX library usage
- [ ] Implement VEX document creation, validation, and manipulation
- [ ] Add streaming support for large document operations
- [ ] Create VEX client abstraction layer
- [ ] Ensure compatibility with existing vexctl output formats

## Tasks

### Task 4.1: VEX Library Integration (2 points)

**Goal**: Integrate with openvex/vex Go libraries for native VEX operations

**Checklist:**
- [ ] Add VEX library dependencies to go.mod
- [ ] Create VEX client wrapper for common operations
- [ ] Implement VEX document creation with native APIs
- [ ] Add VEX document validation and serialization
- [ ] Test compatibility with vexctl output formats

**Code Example - VEX Client Implementation:**
```go
// internal/vex/client.go
package vex

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/openvex/vex/pkg/vex"
    "github.com/rosstaco/vexdoc-mcp/pkg/api"
    "github.com/rosstaco/vexdoc-mcp/internal/logging"
    "github.com/rosstaco/vexdoc-mcp/internal/errors"
)

type Client struct {
    config *api.VEXConfig
    logger logging.Logger
}

func NewClient(config *api.VEXConfig, logger logging.Logger) *Client {
    return &Client{
        config: config,
        logger: logger,
    }
}

func (c *Client) CreateStatement(ctx context.Context, opts *api.CreateOptions) (*api.VEXDocument, error) {
    logger := logging.FromContext(ctx, c.logger)
    logger.Info("Creating VEX statement",
        "product", opts.Product,
        "vulnerability", opts.Vulnerability,
        "status", opts.Status,
    )
    
    // Create VEX document with metadata
    doc := &vex.VEX{
        ID:       c.generateDocumentID(opts),
        Author:   c.getAuthor(opts.Author),
        Version:  1,
        Timestamp: c.getTimestamp(opts.Timestamp),
    }
    
    // Create VEX statement
    statement, err := c.createStatement(opts)
    if err != nil {
        return nil, errors.NewVEXError("Failed to create VEX statement", err)
    }
    
    doc.Statements = append(doc.Statements, statement)
    
    // Validate document if enabled
    if c.config.ValidateOn {
        if err := c.validateDocument(doc); err != nil {
            return nil, errors.NewVEXError("VEX document validation failed", err)
        }
    }
    
    // Convert to API format
    apiDoc, err := c.convertToAPIDocument(doc)
    if err != nil {
        return nil, errors.NewVEXError("Failed to convert VEX document", err)
    }
    
    logger.Info("VEX statement created successfully", "document_id", doc.ID)
    return apiDoc, nil
}

func (c *Client) createStatement(opts *api.CreateOptions) (*vex.Statement, error) {
    // Parse status
    status, err := c.parseStatus(opts.Status)
    if err != nil {
        return nil, fmt.Errorf("invalid status: %w", err)
    }
    
    // Create vulnerability reference
    vulnerability := &vex.Vulnerability{
        Name: opts.Vulnerability,
    }
    
    // Create product reference
    product := &vex.Product{
        Component: &vex.Component{
            ID: opts.Product,
        },
    }
    
    // Create statement
    statement := &vex.Statement{
        Vulnerability: vulnerability,
        Products:      []*vex.Product{product},
        Status:        status,
    }
    
    // Add justification if provided
    if opts.Justification != "" {
        statement.Justification = opts.Justification
    }
    
    return statement, nil
}

func (c *Client) parseStatus(status string) (vex.Status, error) {
    switch status {
    case "not_affected":
        return vex.StatusNotAffected, nil
    case "affected":
        return vex.StatusAffected, nil
    case "fixed":
        return vex.StatusFixed, nil
    case "under_investigation":
        return vex.StatusUnderInvestigation, nil
    default:
        return "", fmt.Errorf("unknown status: %s", status)
    }
}

func (c *Client) validateDocument(doc *vex.VEX) error {
    // Use VEX library validation
    if err := doc.Validate(); err != nil {
        return fmt.Errorf("VEX validation failed: %w", err)
    }
    
    // Additional custom validations
    if len(doc.Statements) == 0 {
        return fmt.Errorf("document must contain at least one statement")
    }
    
    for i, stmt := range doc.Statements {
        if stmt.Vulnerability == nil || stmt.Vulnerability.Name == "" {
            return fmt.Errorf("statement %d: vulnerability name is required", i)
        }
        
        if len(stmt.Products) == 0 {
            return fmt.Errorf("statement %d: at least one product is required", i)
        }
        
        for j, product := range stmt.Products {
            if product.Component == nil || product.Component.ID == "" {
                return fmt.Errorf("statement %d, product %d: component ID is required", i, j)
            }
        }
    }
    
    return nil
}

func (c *Client) convertToAPIDocument(vexDoc *vex.VEX) (*api.VEXDocument, error) {
    // Serialize to JSON and back to ensure compatibility
    jsonBytes, err := json.Marshal(vexDoc)
    if err != nil {
        return nil, fmt.Errorf("marshaling VEX document: %w", err)
    }
    
    var apiDoc api.VEXDocument
    if err := json.Unmarshal(jsonBytes, &apiDoc); err != nil {
        return nil, fmt.Errorf("unmarshaling to API document: %w", err)
    }
    
    return &apiDoc, nil
}

func (c *Client) generateDocumentID(opts *api.CreateOptions) string {
    // Generate a unique document ID based on content
    timestamp := time.Now().Unix()
    return fmt.Sprintf("vex-%s-%s-%d", opts.Product, opts.Vulnerability, timestamp)
}

func (c *Client) getAuthor(author string) string {
    if author != "" {
        return author
    }
    return c.config.DefaultAuthor
}

func (c *Client) getTimestamp(timestamp time.Time) time.Time {
    if timestamp.IsZero() {
        return time.Now()
    }
    return timestamp
}
```

**Dependencies Setup:**
```bash
# Add VEX library dependencies
go get github.com/openvex/vex@latest

# Update go.mod
go mod tidy
```

**Code Example - VEX Types Mapping:**
```go
// pkg/api/vex_types.go
package api

import (
    "time"
    "encoding/json"
)

// VEXDocument represents a VEX document in our API
type VEXDocument struct {
    ID         string           `json:"@id"`
    Context    string           `json:"@context,omitempty"`
    Author     string           `json:"author"`
    Version    int              `json:"version"`
    Timestamp  time.Time        `json:"timestamp"`
    Statements []VEXStatement   `json:"statements"`
}

// VEXStatement represents a single VEX statement
type VEXStatement struct {
    Vulnerability VEXVulnerability `json:"vulnerability"`
    Products      []VEXProduct     `json:"products"`
    Status        string           `json:"status"`
    Justification string           `json:"justification,omitempty"`
    Impact        string           `json:"impact,omitempty"`
}

// VEXVulnerability represents a vulnerability reference
type VEXVulnerability struct {
    Name        string   `json:"name"`
    Description string   `json:"description,omitempty"`
    Aliases     []string `json:"aliases,omitempty"`
}

// VEXProduct represents a product reference
type VEXProduct struct {
    Component VEXComponent `json:"component"`
}

// VEXComponent represents a component reference
type VEXComponent struct {
    ID   string `json:"@id"`
    Name string `json:"name,omitempty"`
}

// JSON Schema for VEX create options
type JSONSchema struct {
    Type       string                 `json:"type"`
    Properties map[string]interface{} `json:"properties"`
    Required   []string               `json:"required"`
}

func CreateVEXStatementSchema() *JSONSchema {
    return &JSONSchema{
        Type: "object",
        Properties: map[string]interface{}{
            "product": map[string]interface{}{
                "type":        "string",
                "description": "Product identifier or component reference",
            },
            "vulnerability": map[string]interface{}{
                "type":        "string",
                "description": "CVE ID or vulnerability identifier",
            },
            "status": map[string]interface{}{
                "type":        "string",
                "enum":        []string{"not_affected", "affected", "fixed", "under_investigation"},
                "description": "VEX status for the vulnerability",
            },
            "justification": map[string]interface{}{
                "type":        "string",
                "description": "Optional justification for the status",
            },
            "author": map[string]interface{}{
                "type":        "string",
                "description": "Optional author override",
            },
        },
        Required: []string{"product", "vulnerability", "status"},
    }
}
```

**Deliverable**: VEX client implementation in `internal/vex/client.go`

---

### Task 4.2: Document Merge Implementation (2 points)

**Goal**: Implement native VEX document merging with conflict resolution

**Checklist:**
- [ ] Create VEX document merge algorithms
- [ ] Implement conflict detection and resolution
- [ ] Add merge options and configuration
- [ ] Support multiple document input formats
- [ ] Ensure deterministic merge results

**Code Example - Document Merging:**
```go
// internal/vex/merge.go
package vex

import (
    "context"
    "encoding/json"
    "fmt"
    "sort"
    "strings"
    
    "github.com/openvex/vex/pkg/vex"
    "github.com/rosstaco/vexdoc-mcp/pkg/api"
    "github.com/rosstaco/vexdoc-mcp/internal/logging"
    "github.com/rosstaco/vexdoc-mcp/internal/errors"
)

func (c *Client) MergeDocuments(ctx context.Context, opts *api.MergeOptions) (*api.VEXDocument, error) {
    logger := logging.FromContext(ctx, c.logger)
    logger.Info("Merging VEX documents", "count", len(opts.Documents))
    
    if len(opts.Documents) < 2 {
        return nil, errors.NewValidationError("At least 2 documents required for merge", nil)
    }
    
    // Parse input documents
    vexDocs, err := c.parseDocuments(opts.Documents)
    if err != nil {
        return nil, errors.NewVEXError("Failed to parse input documents", err)
    }
    
    // Perform merge
    merged, err := c.performMerge(vexDocs, opts)
    if err != nil {
        return nil, errors.NewVEXError("Document merge failed", err)
    }
    
    // Validate merged document
    if c.config.ValidateOn {
        if err := c.validateDocument(merged); err != nil {
            return nil, errors.NewVEXError("Merged document validation failed", err)
        }
    }
    
    // Convert to API format
    apiDoc, err := c.convertToAPIDocument(merged)
    if err != nil {
        return nil, errors.NewVEXError("Failed to convert merged document", err)
    }
    
    logger.Info("VEX documents merged successfully",
        "input_count", len(opts.Documents),
        "output_statements", len(merged.Statements),
    )
    
    return apiDoc, nil
}

func (c *Client) parseDocuments(documents []string) ([]*vex.VEX, error) {
    var vexDocs []*vex.VEX
    
    for i, docStr := range documents {
        var doc vex.VEX
        if err := json.Unmarshal([]byte(docStr), &doc); err != nil {
            return nil, fmt.Errorf("parsing document %d: %w", i, err)
        }
        vexDocs = append(vexDocs, &doc)
    }
    
    return vexDocs, nil
}

func (c *Client) performMerge(docs []*vex.VEX, opts *api.MergeOptions) (*vex.VEX, error) {
    // Create merged document with metadata from first doc
    merged := &vex.VEX{
        ID:        c.generateMergedID(docs, opts.OutputID),
        Author:    c.getAuthor(opts.Author),
        Version:   1,
        Timestamp: time.Now(),
    }
    
    // Collect all statements
    statementMap := make(map[string]*StatementGroup)
    
    for _, doc := range docs {
        for _, stmt := range doc.Statements {
            key := c.createStatementKey(stmt)
            
            if group, exists := statementMap[key]; exists {
                group.Statements = append(group.Statements, stmt)
            } else {
                statementMap[key] = &StatementGroup{
                    Key:        key,
                    Statements: []*vex.Statement{stmt},
                }
            }
        }
    }
    
    // Merge conflicting statements
    for _, group := range statementMap {
        mergedStmt, err := c.mergeStatementGroup(group)
        if err != nil {
            return nil, fmt.Errorf("merging statements for %s: %w", group.Key, err)
        }
        merged.Statements = append(merged.Statements, mergedStmt)
    }
    
    // Sort statements for deterministic output
    sort.Slice(merged.Statements, func(i, j int) bool {
        return c.createStatementKey(merged.Statements[i]) < 
               c.createStatementKey(merged.Statements[j])
    })
    
    return merged, nil
}

type StatementGroup struct {
    Key        string
    Statements []*vex.Statement
}

func (c *Client) createStatementKey(stmt *vex.Statement) string {
    var parts []string
    
    if stmt.Vulnerability != nil {
        parts = append(parts, stmt.Vulnerability.Name)
    }
    
    for _, product := range stmt.Products {
        if product.Component != nil {
            parts = append(parts, product.Component.ID)
        }
    }
    
    return strings.Join(parts, "|")
}

func (c *Client) mergeStatementGroup(group *StatementGroup) (*vex.Statement, error) {
    if len(group.Statements) == 1 {
        return group.Statements[0], nil
    }
    
    // Use first statement as base
    merged := &vex.Statement{
        Vulnerability: group.Statements[0].Vulnerability,
        Products:      group.Statements[0].Products,
    }
    
    // Merge status with priority: fixed > not_affected > under_investigation > affected
    merged.Status = c.getMergedStatus(group.Statements)
    
    // Merge justifications
    merged.Justification = c.getMergedJustification(group.Statements)
    
    return merged, nil
}

func (c *Client) getMergedStatus(statements []*vex.Statement) vex.Status {
    statusPriority := map[vex.Status]int{
        vex.StatusFixed:              4,
        vex.StatusNotAffected:        3,
        vex.StatusUnderInvestigation: 2,
        vex.StatusAffected:           1,
    }
    
    highest := vex.StatusAffected
    highestPriority := 0
    
    for _, stmt := range statements {
        if priority, exists := statusPriority[stmt.Status]; exists && priority > highestPriority {
            highest = stmt.Status
            highestPriority = priority
        }
    }
    
    return highest
}

func (c *Client) getMergedJustification(statements []*vex.Statement) string {
    var justifications []string
    seen := make(map[string]bool)
    
    for _, stmt := range statements {
        if stmt.Justification != "" && !seen[stmt.Justification] {
            justifications = append(justifications, stmt.Justification)
            seen[stmt.Justification] = true
        }
    }
    
    if len(justifications) == 0 {
        return ""
    }
    
    if len(justifications) == 1 {
        return justifications[0]
    }
    
    return fmt.Sprintf("Multiple justifications: %s", strings.Join(justifications, "; "))
}

func (c *Client) generateMergedID(docs []*vex.VEX, outputID string) string {
    if outputID != "" {
        return outputID
    }
    
    // Generate ID based on input document IDs
    var ids []string
    for _, doc := range docs {
        ids = append(ids, doc.ID)
    }
    sort.Strings(ids)
    
    return fmt.Sprintf("merged-%s", strings.Join(ids, "-"))
}
```

**Deliverable**: Document merging implementation in `internal/vex/merge.go`

---

### Task 4.3: Streaming Support (1 point)

**Goal**: Add streaming capabilities for large document operations

**Checklist:**
- [ ] Implement streaming merge for large document sets
- [ ] Add progress reporting for long-running operations
- [ ] Create streaming interfaces for VEX operations
- [ ] Add memory-efficient document processing
- [ ] Test with large document sets

**Code Example - Streaming Implementation:**
```go
// internal/vex/streaming.go
package vex

import (
    "context"
    "fmt"
    "sync"
    
    "github.com/rosstaco/vexdoc-mcp/pkg/api"
    "github.com/rosstaco/vexdoc-mcp/internal/logging"
)

func (c *Client) StreamMerge(ctx context.Context, opts *api.StreamMergeOptions) (<-chan *api.MergeResult, error) {
    logger := logging.FromContext(ctx, c.logger)
    logger.Info("Starting streaming merge", "documents", len(opts.Documents))
    
    resultCh := make(chan *api.MergeResult, 10)
    
    go func() {
        defer close(resultCh)
        c.performStreamingMerge(ctx, opts, resultCh)
    }()
    
    return resultCh, nil
}

func (c *Client) performStreamingMerge(ctx context.Context, opts *api.StreamMergeOptions, resultCh chan<- *api.MergeResult) {
    logger := logging.FromContext(ctx, c.logger)
    
    // Send initial status
    resultCh <- &api.MergeResult{
        Progress: 0.0,
        Status:   "Starting merge operation",
    }
    
    // Parse documents in chunks to manage memory
    chunkSize := opts.ChunkSize
    if chunkSize <= 0 {
        chunkSize = 10 // Default chunk size
    }
    
    totalDocs := len(opts.Documents)
    var allStatements []*StatementGroup
    statementMap := make(map[string]*StatementGroup)
    
    // Process documents in chunks
    for i := 0; i < totalDocs; i += chunkSize {
        select {
        case <-ctx.Done():
            resultCh <- &api.MergeResult{
                Status: "Operation cancelled",
                Error:  ctx.Err(),
            }
            return
        default:
        }
        
        end := i + chunkSize
        if end > totalDocs {
            end = totalDocs
        }
        
        chunk := opts.Documents[i:end]
        
        // Process chunk
        if err := c.processDocumentChunk(chunk, statementMap); err != nil {
            resultCh <- &api.MergeResult{
                Status: fmt.Sprintf("Error processing chunk %d-%d", i, end),
                Error:  err,
            }
            return
        }
        
        progress := float64(end) / float64(totalDocs) * 0.8 // 80% for parsing
        resultCh <- &api.MergeResult{
            Progress: progress,
            Status:   fmt.Sprintf("Processed %d/%d documents", end, totalDocs),
        }
        
        logger.Debug("Processed document chunk", "processed", end, "total", totalDocs)
    }
    
    // Convert map to slice for merging
    for _, group := range statementMap {
        allStatements = append(allStatements, group)
    }
    
    resultCh <- &api.MergeResult{
        Progress: 0.85,
        Status:   "Merging statements",
    }
    
    // Merge statements
    merged, err := c.mergeStatements(ctx, allStatements, opts)
    if err != nil {
        resultCh <- &api.MergeResult{
            Status: "Error merging statements",
            Error:  err,
        }
        return
    }
    
    resultCh <- &api.MergeResult{
        Progress: 0.95,
        Status:   "Validating merged document",
    }
    
    // Validate if enabled
    if c.config.ValidateOn {
        if err := c.validateDocument(merged); err != nil {
            resultCh <- &api.MergeResult{
                Status: "Validation failed",
                Error:  err,
            }
            return
        }
    }
    
    // Convert to API format
    apiDoc, err := c.convertToAPIDocument(merged)
    if err != nil {
        resultCh <- &api.MergeResult{
            Status: "Error converting document",
            Error:  err,
        }
        return
    }
    
    // Send final result
    resultCh <- &api.MergeResult{
        Document: apiDoc,
        Progress: 1.0,
        Status:   "Merge completed successfully",
    }
    
    logger.Info("Streaming merge completed",
        "input_documents", totalDocs,
        "output_statements", len(merged.Statements),
    )
}

func (c *Client) processDocumentChunk(chunk []string, statementMap map[string]*StatementGroup) error {
    for _, docStr := range chunk {
        vexDoc, err := c.parseDocument(docStr)
        if err != nil {
            return fmt.Errorf("parsing document: %w", err)
        }
        
        for _, stmt := range vexDoc.Statements {
            key := c.createStatementKey(stmt)
            
            if group, exists := statementMap[key]; exists {
                group.Statements = append(group.Statements, stmt)
            } else {
                statementMap[key] = &StatementGroup{
                    Key:        key,
                    Statements: []*vex.Statement{stmt},
                }
            }
        }
    }
    
    return nil
}

func (c *Client) parseDocument(docStr string) (*vex.VEX, error) {
    var doc vex.VEX
    if err := json.Unmarshal([]byte(docStr), &doc); err != nil {
        return nil, fmt.Errorf("unmarshaling document: %w", err)
    }
    return &doc, nil
}
```

**Deliverable**: Streaming implementation in `internal/vex/streaming.go`

---

## Phase 4 Deliverables

### 1. VEX Client (`internal/vex/client.go`)
- [ ] Native VEX library integration
- [ ] VEX document creation with validation
- [ ] Type conversion between VEX and API formats
- [ ] Configuration-driven behavior

### 2. Document Merging (`internal/vex/merge.go`)
- [ ] Multi-document merge with conflict resolution
- [ ] Deterministic merge results
- [ ] Status priority handling
- [ ] Justification merging

### 3. Streaming Support (`internal/vex/streaming.go`)
- [ ] Memory-efficient document processing
- [ ] Progress reporting for long operations
- [ ] Chunked document processing
- [ ] Cancellation support

### 4. VEX Type Definitions (`pkg/api/vex_types.go`)
- [ ] Complete VEX document type mapping
- [ ] JSON schema definitions
- [ ] Validation structures

### 5. Integration Tests
- [ ] VEX creation with various inputs
- [ ] Document merging with conflicts
- [ ] Streaming operations with large datasets
- [ ] Error handling and validation

## Success Criteria
- [ ] All VEX operations work without subprocess calls
- [ ] Output format matches vexctl compatibility
- [ ] Streaming operations handle large documents efficiently
- [ ] Memory usage remains stable during operations
- [ ] Performance improvement over subprocess approach

## Dependencies
- **Input**: Phase 3 MCP server framework
- **Output**: Native VEX capabilities ready for tool integration

## Risks & Mitigation
- **Risk**: VEX library API incompatibilities
  - **Mitigation**: Extensive testing with real vexctl examples
- **Risk**: Memory usage with large documents
  - **Mitigation**: Streaming and chunked processing
- **Risk**: Merge algorithm complexity
  - **Mitigation**: Simple, deterministic conflict resolution

## Testing Strategy
```bash
# Unit tests
go test ./internal/vex/...

# Integration tests with real VEX documents
go test -tags=integration ./test/vex/...

# Performance tests
go test -bench=. ./internal/vex/...

# Memory tests
go test -memprofile=mem.prof ./internal/vex/...
```

## Performance Targets
- VEX creation: <10ms (vs ~100ms subprocess)
- Document merge: <100ms for 10 documents
- Streaming: Handle 1000+ documents without memory growth
- Memory: <100MB peak for large operations

## Time Estimate
**5 Story Points** ≈ 2-3 days of focused development

---
**Next**: [Phase 5: Tool Implementation](./phase5-tools.md)
