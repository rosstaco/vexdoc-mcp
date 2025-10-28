# Phase 1: Research & Discovery
**Story Points**: 3 | **Prerequisites**: None | **Next**: [Phase 2](./phase2-foundation.md)

## Overview
Research the feasibility of native Go integration with vexctl and establish the development environment.

## Objectives
- [ ] Validate vexctl Go library availability and APIs
- [ ] Evaluate MCP protocol implementation options  
- [ ] Set up Go development environment
- [ ] Create feasibility report with go/no-go recommendation

## Tasks

### Task 1.1: VEX Library Research (1 point)

**Goal**: Analyze available Go libraries for VEX operations

**Checklist:**
- [ ] Clone `github.com/openvex/vex` repository
- [ ] Examine public APIs in `pkg/` directories
- [ ] Test basic VEX document creation
- [ ] Document available operations (create, merge, validate)
- [ ] Identify any missing functionality vs current Node.js implementation

**Code Example - Testing VEX Library:**
```bash
# Create research directory
mkdir -p research/vex-test
cd research/vex-test

# Initialize Go module
go mod init vex-research

# Add VEX dependency
go get github.com/openvex/vex@latest
```

```go
// main.go - Basic VEX API test
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "time"
    
    "github.com/openvex/vex/pkg/vex"
)

func main() {
    fmt.Println("Testing VEX Go Library APIs...")
    
    // Test 1: Create basic VEX document
    doc := &vex.VEX{
        ID:        "test-vex-001",
        Author:    "research@example.com", 
        Version:   1,
        Timestamp: time.Now(),
    }
    
    // Test 2: Create VEX statement
    statement := &vex.Statement{
        Vulnerability: &vex.Vulnerability{Name: "CVE-2023-1234"},
        Products: []*vex.Product{{
            Component: &vex.Component{ID: "test-product"},
        }},
        Status: vex.StatusNotAffected,
        Justification: "Component not present in build",
    }
    
    doc.Statements = append(doc.Statements, statement)
    
    // Test 3: Serialize to JSON
    jsonBytes, err := json.MarshalIndent(doc, "", "  ")
    if err != nil {
        log.Fatal("Failed to marshal VEX document:", err)
    }
    
    fmt.Printf("Generated VEX document:\n%s\n", string(jsonBytes))
    
    // Test 4: Document API exploration
    testDocumentAPIs(doc)
}

func testDocumentAPIs(doc *vex.VEX) {
    fmt.Println("\n=== API Exploration ===")
    
    // Test validation
    if err := doc.Validate(); err != nil {
        fmt.Printf("Validation failed: %v\n", err)
    } else {
        fmt.Println("✓ Document validation passed")
    }
    
    // Test statement manipulation
    fmt.Printf("Document has %d statements\n", len(doc.Statements))
    
    // Document available methods
    fmt.Println("\nAvailable VEX operations to document:")
    fmt.Println("- Document creation: ✓")
    fmt.Println("- Statement creation: ✓") 
    fmt.Println("- JSON serialization: ✓")
    fmt.Println("- Validation: ✓")
    fmt.Println("- Merge operations: ?") // To be tested
    fmt.Println("- Streaming: ?") // To be tested
}
```

**Deliverable**: `research-vex-apis.md` with findings

---

### Task 1.2: MCP Protocol Research (1 point)

**Goal**: Evaluate options for implementing MCP protocol in Go

**Checklist:**
- [ ] Search for existing Go MCP implementations
- [ ] Review MCP protocol specification
- [ ] Analyze Node.js MCP SDK structure for reference
- [ ] Determine if custom implementation is required
- [ ] Document protocol requirements (JSON-RPC, transports)

**Research Commands:**
```bash
# Search GitHub for Go MCP implementations
gh search repos "mcp golang" --language=go
gh search repos "model context protocol" --language=go

# Clone MCP specification
git clone https://github.com/modelcontextprotocol/specification.git

# Review existing examples
gh search repos "modelcontextprotocol" --language=go
```

**Code Example - MCP Protocol Skeleton:**
```go
// mcp_test.go - Basic MCP protocol structure
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

type MCPError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

func main() {
    fmt.Println("MCP Protocol Structure Test")
    
    // Test request parsing
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
    
    fmt.Printf("Parsed MCP request: %+v\n", req)
    
    // Test response creation
    resp := MCPResponse{
        JSONRPC: "2.0",
        ID:      req.ID,
        Result: map[string]interface{}{
            "tools": []interface{}{},
        },
    }
    
    respJSON, _ := json.MarshalIndent(resp, "", "  ")
    fmt.Printf("MCP response:\n%s\n", string(respJSON))
}
```

**Deliverable**: `research-mcp-protocol.md` with implementation approach

---

### Task 1.3: Environment Setup (1 point)

**Goal**: Establish Go development environment and project structure

**Checklist:**
- [ ] Verify Go 1.21+ installation
- [ ] Configure Go development tools (gopls, gofmt, etc.)
- [ ] Set up project repository structure
- [ ] Initialize Go module with proper naming
- [ ] Configure VS Code/IDE for Go development
- [ ] Set up basic CI/CD skeleton

**Setup Commands:**
```bash
# Verify Go installation
go version

# Create project directory
mkdir -p ~/projects/vexdoc-mcp-go
cd ~/projects/vexdoc-mcp-go

# Initialize Go module
go mod init github.com/rosstaco/vexdoc-mcp

# Create basic project structure
mkdir -p {cmd/server,internal/{mcp,tools,vex},pkg/api,test,docs}

# Create initial files
touch cmd/server/main.go
touch internal/mcp/server.go
touch README.md

# Initialize git repository
git init
git add .
git commit -m "Initial project structure"
```

**Basic Project Structure:**
```
vexdoc-mcp-go/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── mcp/
│   │   └── server.go
│   ├── tools/
│   └── vex/
├── pkg/
│   └── api/
├── test/
├── docs/
├── go.mod
├── go.sum
├── README.md
└── .gitignore
```

**Deliverable**: Working Go development environment

---

## Phase 1 Deliverables

### 1. Feasibility Report (`feasibility-report.md`)
**Required Content:**
- [ ] VEX library capability assessment
- [ ] MCP protocol implementation approach
- [ ] Performance expectations vs Node.js
- [ ] Risk assessment and mitigation strategies
- [ ] Go/No-go recommendation with justification

### 2. Development Environment
- [ ] Go module initialized
- [ ] Basic project structure created
- [ ] Development tools configured
- [ ] Initial repository setup

### 3. Research Artifacts
- [ ] `research-vex-apis.md` - VEX library analysis
- [ ] `research-mcp-protocol.md` - MCP implementation approach
- [ ] Working code examples for both VEX and MCP

## Success Criteria
- [ ] All VEX operations (create, merge) are possible with Go libraries
- [ ] MCP protocol can be implemented (custom or existing library)
- [ ] Development environment is fully functional
- [ ] Clear implementation path identified
- [ ] Stakeholder approval to proceed to Phase 2

## Risks & Mitigation
- **Risk**: VEX Go APIs insufficient
  - **Mitigation**: Document gaps, plan subprocess fallback
- **Risk**: MCP Go ecosystem immature  
  - **Mitigation**: Plan custom implementation from specification

## Time Estimate
**3 Story Points** ≈ 1-2 days with proper focus

---
**Next**: [Phase 2: Project Foundation](./phase2-foundation.md)
