# Phase 3: MCP Protocol Core
**Story Points**: 8 | **Prerequisites**: [Phase 2](./phase2-foundation.md) | **Next**: [Phase 4](./phase4-vex-integration.md)

## Overview
Implement the Model Context Protocol (MCP) server core with JSON-RPC handling and multiple transport layers (stdio, HTTP, streaming).

## Objectives
- [ ] Implement MCP JSON-RPC protocol handling
- [ ] Create transport abstraction for stdio, HTTP, and streaming
- [ ] Add capability negotiation and tool registration
- [ ] Implement request/response lifecycle management
- [ ] Add comprehensive error handling for protocol operations

## Tasks

### Task 3.1: MCP JSON-RPC Core (3 points)

**Goal**: Implement the core MCP protocol handling with JSON-RPC 2.0

**Checklist:**
- [ ] Create JSON-RPC message parsing and validation
- [ ] Implement request routing and method dispatch
- [ ] Add response serialization and error handling
- [ ] Create protocol state management
- [ ] Add request/response correlation

**Code Example - Protocol Handler:**
```go
// internal/mcp/protocol.go
package mcp

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    
    "github.com/rosstaco/vexdoc-mcp/pkg/api"
    "github.com/rosstaco/vexdoc-mcp/internal/logging"
    "github.com/rosstaco/vexdoc-mcp/internal/errors"
)

type ProtocolHandler struct {
    logger   logging.Logger
    handlers map[string]RequestHandler
    tools    map[string]api.Tool
    mu       sync.RWMutex
    
    // Server capabilities
    capabilities *api.ServerCapabilities
}

type RequestHandler func(ctx context.Context, params json.RawMessage) (interface{}, error)

func NewProtocolHandler(logger logging.Logger) *ProtocolHandler {
    ph := &ProtocolHandler{
        logger:   logger,
        handlers: make(map[string]RequestHandler),
        tools:    make(map[string]api.Tool),
        capabilities: &api.ServerCapabilities{
            Tools: &api.ToolCapabilities{},
        },
    }
    
    // Register core MCP methods
    ph.registerCoreHandlers()
    
    return ph
}

func (ph *ProtocolHandler) registerCoreHandlers() {
    ph.handlers["initialize"] = ph.handleInitialize
    ph.handlers["tools/list"] = ph.handleToolsList
    ph.handlers["tools/call"] = ph.handleToolsCall
}

func (ph *ProtocolHandler) HandleRequest(ctx context.Context, req *api.Request) *api.Response {
    logger := logging.FromContext(ctx, ph.logger).With(
        "method", req.Method,
        "request_id", req.ID,
    )
    
    logger.Debug("Processing MCP request")
    
    response := &api.Response{
        JSONRPC: "2.0",
        ID:      req.ID,
    }
    
    // Find handler for method
    ph.mu.RLock()
    handler, exists := ph.handlers[req.Method]
    ph.mu.RUnlock()
    
    if !exists {
        logger.Warn("Unknown method requested")
        response.Error = &api.MCPError{
            Code:    -32601,
            Message: fmt.Sprintf("Method not found: %s", req.Method),
        }
        return response
    }
    
    // Parse parameters
    var params json.RawMessage
    if req.Params != nil {
        paramBytes, err := json.Marshal(req.Params)
        if err != nil {
            logger.Error("Failed to marshal parameters", "error", err)
            response.Error = &api.MCPError{
                Code:    -32602,
                Message: "Invalid params",
                Data:    err.Error(),
            }
            return response
        }
        params = paramBytes
    }
    
    // Execute handler
    result, err := handler(ctx, params)
    if err != nil {
        logger.Error("Handler execution failed", "error", err)
        
        var appErr *errors.AppError
        if errors.As(err, &appErr) {
            response.Error = &api.MCPError{
                Code:    int64(appErr.Code),
                Message: appErr.Message,
                Data:    appErr.Type,
            }
        } else {
            response.Error = &api.MCPError{
                Code:    -32603,
                Message: "Internal error",
                Data:    err.Error(),
            }
        }
        return response
    }
    
    response.Result = result
    logger.Debug("Request processed successfully")
    
    return response
}

func (ph *ProtocolHandler) handleInitialize(ctx context.Context, params json.RawMessage) (interface{}, error) {
    var initParams struct {
        ProtocolVersion string                 `json:"protocolVersion"`
        Capabilities    map[string]interface{} `json:"capabilities"`
        ClientInfo      struct {
            Name    string `json:"name"`
            Version string `json:"version"`
        } `json:"clientInfo"`
    }
    
    if len(params) > 0 {
        if err := json.Unmarshal(params, &initParams); err != nil {
            return nil, errors.NewValidationError("Invalid initialize parameters", err)
        }
    }
    
    // Validate protocol version
    if initParams.ProtocolVersion != "" && initParams.ProtocolVersion != "2024-11-05" {
        return nil, errors.NewValidationError(
            fmt.Sprintf("Unsupported protocol version: %s", initParams.ProtocolVersion),
            nil,
        )
    }
    
    ph.logger.Info("Client initialized",
        "client_name", initParams.ClientInfo.Name,
        "client_version", initParams.ClientInfo.Version,
        "protocol_version", initParams.ProtocolVersion,
    )
    
    return map[string]interface{}{
        "protocolVersion": "2024-11-05",
        "serverInfo": map[string]interface{}{
            "name":    "vexdoc-mcp-server",
            "version": "1.0.0",
        },
        "capabilities": ph.capabilities,
    }, nil
}

func (ph *ProtocolHandler) handleToolsList(ctx context.Context, params json.RawMessage) (interface{}, error) {
    ph.mu.RLock()
    defer ph.mu.RUnlock()
    
    tools := make([]map[string]interface{}, 0, len(ph.tools))
    
    for _, tool := range ph.tools {
        toolInfo := map[string]interface{}{
            "name":        tool.Name(),
            "description": tool.Description(),
            "inputSchema": tool.InputSchema(),
        }
        tools = append(tools, toolInfo)
    }
    
    return map[string]interface{}{
        "tools": tools,
    }, nil
}

func (ph *ProtocolHandler) handleToolsCall(ctx context.Context, params json.RawMessage) (interface{}, error) {
    var callParams struct {
        Name      string                 `json:"name"`
        Arguments map[string]interface{} `json:"arguments"`
    }
    
    if err := json.Unmarshal(params, &callParams); err != nil {
        return nil, errors.NewValidationError("Invalid tool call parameters", err)
    }
    
    ph.mu.RLock()
    tool, exists := ph.tools[callParams.Name]
    ph.mu.RUnlock()
    
    if !exists {
        return nil, errors.NewValidationError(
            fmt.Sprintf("Tool not found: %s", callParams.Name),
            nil,
        )
    }
    
    // Add tool name to context for logging
    ctx = logging.WithToolName(ctx, callParams.Name)
    
    // Execute tool
    result, err := tool.Execute(ctx, callParams.Arguments)
    if err != nil {
        return nil, errors.NewVEXError(
            fmt.Sprintf("Tool execution failed: %s", callParams.Name),
            err,
        )
    }
    
    return result, nil
}

// Tool registration
func (ph *ProtocolHandler) RegisterTool(tool api.Tool) error {
    ph.mu.Lock()
    defer ph.mu.Unlock()
    
    name := tool.Name()
    if _, exists := ph.tools[name]; exists {
        return fmt.Errorf("tool already registered: %s", name)
    }
    
    ph.tools[name] = tool
    ph.logger.Info("Tool registered", "tool_name", name)
    
    return nil
}
```

**Deliverable**: Core protocol handler in `internal/mcp/protocol.go`

---

### Task 3.2: Transport Layer Implementation (3 points)

**Goal**: Create transport abstraction supporting stdio, HTTP, and streaming

**Checklist:**
- [ ] Implement transport interface and abstraction
- [ ] Create stdio transport for command-line usage
- [ ] Implement HTTP transport for web integration
- [ ] Add streaming HTTP transport for real-time communication
- [ ] Create transport factory and configuration

**Code Example - Transport Interface:**
```go
// internal/mcp/transport.go
package mcp

import (
    "context"
    "bufio"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "sync"
    
    "github.com/rosstaco/vexdoc-mcp/pkg/api"
    "github.com/rosstaco/vexdoc-mcp/internal/logging"
)

// StdioTransport handles communication via stdin/stdout
type StdioTransport struct {
    reader io.Reader
    writer io.Writer
    logger logging.Logger
    
    requestCh  chan *api.Request
    responseCh chan *api.Response
    stopCh     chan struct{}
    wg         sync.WaitGroup
}

func NewStdioTransport(reader io.Reader, writer io.Writer, logger logging.Logger) *StdioTransport {
    return &StdioTransport{
        reader:     reader,
        writer:     writer,
        logger:     logger,
        requestCh:  make(chan *api.Request, 10),
        responseCh: make(chan *api.Response, 10),
        stopCh:     make(chan struct{}),
    }
}

func (t *StdioTransport) Start(ctx context.Context) error {
    t.logger.Info("Starting stdio transport")
    
    // Start reader goroutine
    t.wg.Add(1)
    go t.readLoop(ctx)
    
    // Start writer goroutine
    t.wg.Add(1)
    go t.writeLoop(ctx)
    
    return nil
}

func (t *StdioTransport) Stop() error {
    t.logger.Info("Stopping stdio transport")
    close(t.stopCh)
    t.wg.Wait()
    return nil
}

func (t *StdioTransport) Receive() <-chan *api.Request {
    return t.requestCh
}

func (t *StdioTransport) Send(response *api.Response) error {
    select {
    case t.responseCh <- response:
        return nil
    case <-t.stopCh:
        return fmt.Errorf("transport stopped")
    }
}

func (t *StdioTransport) readLoop(ctx context.Context) {
    defer t.wg.Done()
    defer close(t.requestCh)
    
    scanner := bufio.NewScanner(t.reader)
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-t.stopCh:
            return
        default:
            if !scanner.Scan() {
                if err := scanner.Err(); err != nil {
                    t.logger.Error("Stdin read error", "error", err)
                }
                return
            }
            
            line := scanner.Text()
            if line == "" {
                continue
            }
            
            var req api.Request
            if err := json.Unmarshal([]byte(line), &req); err != nil {
                t.logger.Error("Failed to parse request", "error", err, "line", line)
                continue
            }
            
            select {
            case t.requestCh <- &req:
            case <-ctx.Done():
                return
            case <-t.stopCh:
                return
            }
        }
    }
}

func (t *StdioTransport) writeLoop(ctx context.Context) {
    defer t.wg.Done()
    
    encoder := json.NewEncoder(t.writer)
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-t.stopCh:
            return
        case response := <-t.responseCh:
            if err := encoder.Encode(response); err != nil {
                t.logger.Error("Failed to write response", "error", err)
            }
        }
    }
}
```

**Code Example - HTTP Transport:**
```go
// HTTPTransport handles communication via HTTP
type HTTPTransport struct {
    server *http.Server
    logger logging.Logger
    
    requestCh  chan *api.Request
    responseCh map[string]chan *api.Response
    mu         sync.RWMutex
}

func NewHTTPTransport(port int, logger logging.Logger) *HTTPTransport {
    transport := &HTTPTransport{
        logger:     logger,
        requestCh:  make(chan *api.Request, 100),
        responseCh: make(map[string]chan *api.Response),
    }
    
    mux := http.NewServeMux()
    mux.HandleFunc("/mcp", transport.handleMCP)
    
    transport.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: mux,
    }
    
    return transport
}

func (t *HTTPTransport) Start(ctx context.Context) error {
    t.logger.Info("Starting HTTP transport", "addr", t.server.Addr)
    
    go func() {
        if err := t.server.ListenAndServe(); err != http.ErrServerClosed {
            t.logger.Error("HTTP server error", "error", err)
        }
    }()
    
    return nil
}

func (t *HTTPTransport) Stop() error {
    t.logger.Info("Stopping HTTP transport")
    return t.server.Shutdown(context.Background())
}

func (t *HTTPTransport) Receive() <-chan *api.Request {
    return t.requestCh
}

func (t *HTTPTransport) Send(response *api.Response) error {
    requestID := fmt.Sprintf("%v", response.ID)
    
    t.mu.RLock()
    ch, exists := t.responseCh[requestID]
    t.mu.RUnlock()
    
    if !exists {
        return fmt.Errorf("no response channel for request ID: %s", requestID)
    }
    
    select {
    case ch <- response:
        return nil
    default:
        return fmt.Errorf("response channel full for request ID: %s", requestID)
    }
}

func (t *HTTPTransport) handleMCP(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var req api.Request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // Create response channel for this request
    requestID := fmt.Sprintf("%v", req.ID)
    responseCh := make(chan *api.Response, 1)
    
    t.mu.Lock()
    t.responseCh[requestID] = responseCh
    t.mu.Unlock()
    
    // Cleanup response channel
    defer func() {
        t.mu.Lock()
        delete(t.responseCh, requestID)
        t.mu.Unlock()
        close(responseCh)
    }()
    
    // Send request for processing
    select {
    case t.requestCh <- &req:
    case <-r.Context().Done():
        http.Error(w, "Request cancelled", http.StatusRequestTimeout)
        return
    }
    
    // Wait for response
    select {
    case response := <-responseCh:
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    case <-r.Context().Done():
        http.Error(w, "Request timeout", http.StatusRequestTimeout)
    }
}
```

**Code Example - Transport Factory:**
```go
// internal/mcp/factory.go
package mcp

import (
    "fmt"
    "io"
    "os"
    
    "github.com/rosstaco/vexdoc-mcp/pkg/api"
    "github.com/rosstaco/vexdoc-mcp/internal/logging"
)

func CreateTransport(config api.TransportConfig, logger logging.Logger) (api.Transport, error) {
    switch config.Type {
    case "stdio":
        return NewStdioTransport(os.Stdin, os.Stdout, logger), nil
        
    case "http":
        return NewHTTPTransport(config.Port, logger), nil
        
    case "streaming":
        return NewStreamingHTTPTransport(config.Port, logger), nil
        
    default:
        return nil, fmt.Errorf("unsupported transport type: %s", config.Type)
    }
}
```

**Deliverable**: Complete transport layer in `internal/mcp/transport.go`

---

### Task 3.3: MCP Server Implementation (2 points)

**Goal**: Integrate protocol handler with transport layer into complete MCP server

**Checklist:**
- [ ] Create main MCP server structure
- [ ] Integrate protocol handler with transport layer
- [ ] Implement server lifecycle management (start/stop)
- [ ] Add graceful shutdown handling
- [ ] Create server factory for different configurations

**Code Example - MCP Server:**
```go
// internal/mcp/server.go
package mcp

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "sync"
    "syscall"
    
    "github.com/rosstaco/vexdoc-mcp/pkg/api"
    "github.com/rosstaco/vexdoc-mcp/internal/logging"
)

type Server struct {
    config    *api.ServerConfig
    logger    logging.Logger
    transport api.Transport
    protocol  *ProtocolHandler
    
    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup
}

func NewServer(config *api.ServerConfig, logger logging.Logger) (*Server, error) {
    // Create transport
    transport, err := CreateTransport(config.Transport, logger)
    if err != nil {
        return nil, fmt.Errorf("creating transport: %w", err)
    }
    
    // Create protocol handler
    protocol := NewProtocolHandler(logger)
    
    ctx, cancel := context.WithCancel(context.Background())
    
    return &Server{
        config:    config,
        logger:    logger,
        transport: transport,
        protocol:  protocol,
        ctx:       ctx,
        cancel:    cancel,
    }, nil
}

func (s *Server) RegisterTool(tool api.Tool) error {
    return s.protocol.RegisterTool(tool)
}

func (s *Server) Start() error {
    s.logger.Info("Starting MCP server",
        "name", s.config.Name,
        "version", s.config.Version,
        "transport", s.config.Transport.Type,
    )
    
    // Start transport
    if err := s.transport.Start(s.ctx); err != nil {
        return fmt.Errorf("starting transport: %w", err)
    }
    
    // Start request processing loop
    s.wg.Add(1)
    go s.processRequests()
    
    // Set up signal handling for graceful shutdown
    s.wg.Add(1)
    go s.handleSignals()
    
    return nil
}

func (s *Server) Stop() error {
    s.logger.Info("Stopping MCP server")
    
    // Cancel context to stop all operations
    s.cancel()
    
    // Stop transport
    if err := s.transport.Stop(); err != nil {
        s.logger.Error("Error stopping transport", "error", err)
    }
    
    // Wait for all goroutines to finish
    s.wg.Wait()
    
    s.logger.Info("MCP server stopped")
    return nil
}

func (s *Server) Wait() {
    s.wg.Wait()
}

func (s *Server) processRequests() {
    defer s.wg.Done()
    
    requestCh := s.transport.Receive()
    
    for {
        select {
        case <-s.ctx.Done():
            return
            
        case req, ok := <-requestCh:
            if !ok {
                s.logger.Info("Request channel closed")
                return
            }
            
            // Process request in separate goroutine to avoid blocking
            go s.handleRequest(req)
        }
    }
}

func (s *Server) handleRequest(req *api.Request) {
    // Add request ID to context for tracing
    ctx := logging.WithRequestID(s.ctx, fmt.Sprintf("%v", req.ID))
    
    logger := logging.FromContext(ctx, s.logger)
    logger.Debug("Processing request", "method", req.Method)
    
    // Process request through protocol handler
    response := s.protocol.HandleRequest(ctx, req)
    
    // Send response
    if err := s.transport.Send(response); err != nil {
        logger.Error("Failed to send response", "error", err)
    } else {
        logger.Debug("Response sent successfully")
    }
}

func (s *Server) handleSignals() {
    defer s.wg.Done()
    
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    
    select {
    case sig := <-sigCh:
        s.logger.Info("Received signal, shutting down", "signal", sig)
        s.cancel()
    case <-s.ctx.Done():
        return
    }
}
```

**Code Example - Server Factory:**
```go
// cmd/server/main.go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/rosstaco/vexdoc-mcp/internal/config"
    "github.com/rosstaco/vexdoc-mcp/internal/logging"
    "github.com/rosstaco/vexdoc-mcp/internal/mcp"
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
    
    // Create MCP server
    server, err := mcp.NewServer(cfg, logger)
    if err != nil {
        logger.Error("Failed to create server", "error", err)
        os.Exit(1)
    }
    
    // TODO: Register tools (will be implemented in Phase 5)
    
    // Start server
    if err := server.Start(); err != nil {
        logger.Error("Failed to start server", "error", err)
        os.Exit(1)
    }
    
    // Wait for server to finish
    server.Wait()
}
```

**Deliverable**: Complete MCP server in `internal/mcp/server.go` and `cmd/server/main.go`

---

## Phase 3 Deliverables

### 1. MCP Protocol Core (`internal/mcp/protocol.go`)
- [ ] JSON-RPC 2.0 request/response handling
- [ ] Method routing and dispatch
- [ ] Error handling and validation
- [ ] Tool registration and execution
- [ ] Capability negotiation

### 2. Transport Layer (`internal/mcp/transport.go`)
- [ ] Transport interface abstraction
- [ ] Stdio transport implementation
- [ ] HTTP transport implementation
- [ ] Streaming HTTP transport (basic)
- [ ] Transport factory

### 3. MCP Server (`internal/mcp/server.go`)
- [ ] Server lifecycle management
- [ ] Request processing pipeline
- [ ] Graceful shutdown handling
- [ ] Signal handling
- [ ] Integration of protocol and transport

### 4. Server Entry Point (`cmd/server/main.go`)
- [ ] Configuration loading
- [ ] Logger initialization
- [ ] Server creation and startup
- [ ] Basic CLI interface

### 5. Integration Tests
- [ ] Protocol handler unit tests
- [ ] Transport layer tests
- [ ] End-to-end server tests
- [ ] Error handling validation

## Success Criteria
- [ ] MCP server starts successfully with all transport types
- [ ] Protocol correctly handles initialize, tools/list, tools/call methods
- [ ] Error responses follow MCP specification
- [ ] Server shuts down gracefully on signals
- [ ] All tests pass with good coverage (>80%)

## Dependencies
- **Input**: Phase 2 foundation (interfaces, configuration, logging)
- **Output**: Working MCP server ready for tool integration

## Risks & Mitigation
- **Risk**: MCP protocol complexity underestimated
  - **Mitigation**: Start with basic methods, iterate to full compliance
- **Risk**: Transport implementation bugs
  - **Mitigation**: Comprehensive testing, especially for edge cases
- **Risk**: Performance issues with request handling
  - **Mitigation**: Benchmark concurrent request processing

## Testing Strategy
```bash
# Unit tests
go test ./internal/mcp/...

# Integration tests  
go test -tags=integration ./test/integration/...

# Manual testing with MCP client
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | go run cmd/server/main.go
```

## Time Estimate
**8 Story Points** ≈ 3-4 days of focused development

---
**Next**: [Phase 4: VEX Native Integration](./phase4-vex-integration.md)
