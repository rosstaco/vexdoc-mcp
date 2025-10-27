# Phase 2: Project Foundation
**Story Points**: 5 | **Prerequisites**: [Phase 1](./phase1-research.md) | **Next**: [Phase 3](./phase3-mcp-core.md)

## Overview
Establish the core project architecture, interfaces, and foundational code structure based on Phase 1 research findings.

## Objectives
- [ ] Create comprehensive project structure
- [ ] Define core interfaces and type system
- [ ] Implement configuration management
- [ ] Set up logging and error handling
- [ ] Establish testing framework

## Tasks

### Task 2.1: Core Architecture & Interfaces (2 points)

**Goal**: Define the foundational interfaces and architecture patterns

**Checklist:**
- [ ] Create core interface definitions
- [ ] Implement configuration system
- [ ] Set up structured logging
- [ ] Define error handling patterns
- [ ] Create type definitions for VEX/MCP integration

**Code Example - Core Interfaces:**
```go
// pkg/api/interfaces.go
package api

import (
    "context"
    "io"
)

// Tool represents an MCP tool that can be executed
type Tool interface {
    Name() string
    Description() string
    InputSchema() *JSONSchema
    Execute(ctx context.Context, args map[string]interface{}) (*ToolResponse, error)
}

// StreamingTool extends Tool with streaming capabilities
type StreamingTool interface {
    Tool
    StreamExecute(ctx context.Context, args map[string]interface{}) (<-chan *ToolResponse, error)
}

// VEXClient handles all VEX document operations
type VEXClient interface {
    CreateStatement(ctx context.Context, opts *CreateOptions) (*VEXDocument, error)
    MergeDocuments(ctx context.Context, opts *MergeOptions) (*VEXDocument, error)
    ValidateDocument(ctx context.Context, doc *VEXDocument) error
    
    // Streaming operations
    StreamMerge(ctx context.Context, opts *StreamMergeOptions) (<-chan *MergeResult, error)
}

// MCPServer handles Model Context Protocol communication
type MCPServer interface {
    Start(ctx context.Context) error
    Stop() error
    RegisterTool(tool Tool) error
    SetCapabilities(caps *ServerCapabilities) error
}

// Transport abstracts different communication methods (stdio, HTTP)
type Transport interface {
    Start(ctx context.Context) error
    Stop() error
    Send(response *Response) error
    Receive() (<-chan *Request, error)
}
```

**Code Example - Type Definitions:**
```go
// pkg/api/types.go
package api

import (
    "time"
    "encoding/json"
)

// MCP Protocol Types
type Request struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

type Response struct {
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

// Tool Response Types
type ToolResponse struct {
    Content []Content `json:"content"`
    IsError bool      `json:"isError,omitempty"`
}

type Content struct {
    Type string `json:"type"`
    Text string `json:"text"`
    Data string `json:"data,omitempty"`
}

// VEX Operation Types
type CreateOptions struct {
    Product       string    `json:"product"`
    Vulnerability string    `json:"vulnerability"`
    Status        string    `json:"status"`
    Justification string    `json:"justification,omitempty"`
    Author        string    `json:"author,omitempty"`
    Timestamp     time.Time `json:"timestamp,omitempty"`
}

type MergeOptions struct {
    Documents []string `json:"documents"`
    OutputID  string   `json:"output_id,omitempty"`
    Author    string   `json:"author,omitempty"`
}

type StreamMergeOptions struct {
    MergeOptions
    ChunkSize int `json:"chunk_size,omitempty"`
}

type MergeResult struct {
    Document *VEXDocument `json:"document,omitempty"`
    Progress float64      `json:"progress"`
    Status   string       `json:"status"`
    Error    error        `json:"error,omitempty"`
}

// Server Configuration
type ServerConfig struct {
    Name         string              `json:"name"`
    Version      string              `json:"version"`
    Transport    TransportConfig     `json:"transport"`
    VEX          VEXConfig          `json:"vex"`
    Logging      LoggingConfig      `json:"logging"`
}

type TransportConfig struct {
    Type string `json:"type"` // "stdio", "http", "streaming"
    Port int    `json:"port,omitempty"`
}

type VEXConfig struct {
    DefaultAuthor string `json:"default_author"`
    ValidateOn    bool   `json:"validate_on_create"`
}

type LoggingConfig struct {
    Level  string `json:"level"`
    Format string `json:"format"`
}
```

**Deliverable**: Complete interface definitions in `pkg/api/`

---

### Task 2.2: Configuration Management (1 point)

**Goal**: Implement flexible configuration system supporting multiple sources

**Checklist:**
- [ ] Create configuration structure
- [ ] Support environment variables
- [ ] Support config file loading
- [ ] Implement configuration validation
- [ ] Add configuration defaults

**Code Example - Configuration System:**
```go
// internal/config/config.go
package config

import (
    "encoding/json"
    "fmt"
    "os"
    "strconv"
    "strings"
    
    "github.com/rosstaco/vexdoc-mcp-go/pkg/api"
)

// Load loads configuration from multiple sources
func Load() (*api.ServerConfig, error) {
    config := defaultConfig()
    
    // Load from environment variables
    if err := loadFromEnv(config); err != nil {
        return nil, fmt.Errorf("loading from environment: %w", err)
    }
    
    // Load from config file if specified
    if configFile := os.Getenv("CONFIG_FILE"); configFile != "" {
        if err := loadFromFile(config, configFile); err != nil {
            return nil, fmt.Errorf("loading from file %s: %w", configFile, err)
        }
    }
    
    // Validate configuration
    if err := validate(config); err != nil {
        return nil, fmt.Errorf("configuration validation: %w", err)
    }
    
    return config, nil
}

func defaultConfig() *api.ServerConfig {
    return &api.ServerConfig{
        Name:    "vexdoc-mcp-server",
        Version: "1.0.0",
        Transport: api.TransportConfig{
            Type: "stdio",
            Port: 3000,
        },
        VEX: api.VEXConfig{
            DefaultAuthor: "vexdoc-mcp-server",
            ValidateOn:    true,
        },
        Logging: api.LoggingConfig{
            Level:  "info",
            Format: "json",
        },
    }
}

func loadFromEnv(config *api.ServerConfig) error {
    envMap := map[string]func(string) error{
        "SERVER_NAME":      func(v string) error { config.Name = v; return nil },
        "SERVER_VERSION":   func(v string) error { config.Version = v; return nil },
        "TRANSPORT_TYPE":   func(v string) error { config.Transport.Type = v; return nil },
        "TRANSPORT_PORT":   func(v string) error { 
            port, err := strconv.Atoi(v)
            if err != nil {
                return fmt.Errorf("invalid port: %w", err)
            }
            config.Transport.Port = port
            return nil
        },
        "VEX_DEFAULT_AUTHOR": func(v string) error { config.VEX.DefaultAuthor = v; return nil },
        "VEX_VALIDATE":       func(v string) error { 
            config.VEX.ValidateOn = strings.ToLower(v) == "true"
            return nil
        },
        "LOG_LEVEL":          func(v string) error { config.Logging.Level = v; return nil },
        "LOG_FORMAT":         func(v string) error { config.Logging.Format = v; return nil },
    }
    
    for envVar, setter := range envMap {
        if value := os.Getenv(envVar); value != "" {
            if err := setter(value); err != nil {
                return fmt.Errorf("setting %s: %w", envVar, err)
            }
        }
    }
    
    return nil
}

func loadFromFile(config *api.ServerConfig, filename string) error {
    data, err := os.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("reading config file: %w", err)
    }
    
    if err := json.Unmarshal(data, config); err != nil {
        return fmt.Errorf("parsing config file: %w", err)
    }
    
    return nil
}

func validate(config *api.ServerConfig) error {
    if config.Name == "" {
        return fmt.Errorf("server name is required")
    }
    
    validTransports := map[string]bool{
        "stdio": true, "http": true, "streaming": true,
    }
    if !validTransports[config.Transport.Type] {
        return fmt.Errorf("invalid transport type: %s", config.Transport.Type)
    }
    
    if config.Transport.Type != "stdio" && config.Transport.Port <= 0 {
        return fmt.Errorf("valid port required for %s transport", config.Transport.Type)
    }
    
    validLogLevels := map[string]bool{
        "debug": true, "info": true, "warn": true, "error": true,
    }
    if !validLogLevels[config.Logging.Level] {
        return fmt.Errorf("invalid log level: %s", config.Logging.Level)
    }
    
    return nil
}
```

**Deliverable**: Configuration system in `internal/config/`

---

### Task 2.3: Logging & Error Handling (1 point)

**Goal**: Implement structured logging and consistent error handling

**Checklist:**
- [ ] Set up structured logging (JSON format)
- [ ] Create logging interfaces and implementations
- [ ] Define error types and handling patterns
- [ ] Add request tracing support
- [ ] Implement log level controls

**Code Example - Logging System:**
```go
// internal/logging/logger.go
package logging

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log/slog"
    "os"
    "time"
)

type Logger interface {
    Debug(msg string, args ...interface{})
    Info(msg string, args ...interface{})
    Warn(msg string, args ...interface{})
    Error(msg string, args ...interface{})
    With(args ...interface{}) Logger
}

type contextKey string

const (
    RequestIDKey contextKey = "request_id"
    ToolNameKey  contextKey = "tool_name"
)

type structuredLogger struct {
    logger *slog.Logger
}

func New(level string, format string, output io.Writer) (Logger, error) {
    var slogLevel slog.Level
    switch level {
    case "debug":
        slogLevel = slog.LevelDebug
    case "info":
        slogLevel = slog.LevelInfo
    case "warn":
        slogLevel = slog.LevelWarn
    case "error":
        slogLevel = slog.LevelError
    default:
        return nil, fmt.Errorf("invalid log level: %s", level)
    }
    
    opts := &slog.HandlerOptions{
        Level: slogLevel,
    }
    
    var handler slog.Handler
    if format == "json" {
        handler = slog.NewJSONHandler(output, opts)
    } else {
        handler = slog.NewTextHandler(output, opts)
    }
    
    return &structuredLogger{
        logger: slog.New(handler),
    }, nil
}

func (l *structuredLogger) Debug(msg string, args ...interface{}) {
    l.logger.Debug(msg, args...)
}

func (l *structuredLogger) Info(msg string, args ...interface{}) {
    l.logger.Info(msg, args...)
}

func (l *structuredLogger) Warn(msg string, args ...interface{}) {
    l.logger.Warn(msg, args...)
}

func (l *structuredLogger) Error(msg string, args ...interface{}) {
    l.logger.Error(msg, args...)
}

func (l *structuredLogger) With(args ...interface{}) Logger {
    return &structuredLogger{
        logger: l.logger.With(args...),
    }
}

// Context helpers
func WithRequestID(ctx context.Context, requestID string) context.Context {
    return context.WithValue(ctx, RequestIDKey, requestID)
}

func WithToolName(ctx context.Context, toolName string) context.Context {
    return context.WithValue(ctx, ToolNameKey, toolName)
}

func FromContext(ctx context.Context, base Logger) Logger {
    logger := base
    
    if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
        logger = logger.With("request_id", requestID)
    }
    
    if toolName, ok := ctx.Value(ToolNameKey).(string); ok {
        logger = logger.With("tool", toolName)
    }
    
    return logger
}
```

**Code Example - Error Types:**
```go
// internal/errors/errors.go
package errors

import (
    "fmt"
    "net/http"
)

// Error types for different categories
type ErrorType string

const (
    ValidationError ErrorType = "validation"
    ConfigError     ErrorType = "config"
    VEXError        ErrorType = "vex"
    MCPError        ErrorType = "mcp"
    InternalError   ErrorType = "internal"
)

// AppError represents application-specific errors
type AppError struct {
    Type    ErrorType `json:"type"`
    Message string    `json:"message"`
    Code    int       `json:"code"`
    Cause   error     `json:"cause,omitempty"`
}

func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Cause
}

// Error constructors
func NewValidationError(message string, cause error) *AppError {
    return &AppError{
        Type:    ValidationError,
        Message: message,
        Code:    http.StatusBadRequest,
        Cause:   cause,
    }
}

func NewVEXError(message string, cause error) *AppError {
    return &AppError{
        Type:    VEXError,
        Message: message,
        Code:    http.StatusUnprocessableEntity,
        Cause:   cause,
    }
}

func NewMCPError(message string, cause error) *AppError {
    return &AppError{
        Type:    MCPError,
        Message: message,
        Code:    http.StatusBadRequest,
        Cause:   cause,
    }
}

func NewInternalError(message string, cause error) *AppError {
    return &AppError{
        Type:    InternalError,
        Message: message,
        Code:    http.StatusInternalServerError,
        Cause:   cause,
    }
}
```

**Deliverable**: Logging and error handling in `internal/logging/` and `internal/errors/`

---

### Task 2.4: Testing Framework Setup (1 point)

**Goal**: Establish comprehensive testing framework and patterns

**Checklist:**
- [ ] Set up unit testing structure
- [ ] Create test helpers and mocks
- [ ] Configure test coverage reporting
- [ ] Set up integration test framework
- [ ] Create testing utilities for MCP/VEX

**Code Example - Test Helpers:**
```go
// test/helpers/mcp.go
package helpers

import (
    "context"
    "encoding/json"
    "testing"
    
    "github.com/rosstaco/vexdoc-mcp-go/pkg/api"
)

// MockVEXClient for testing
type MockVEXClient struct {
    CreateStatementFunc func(context.Context, *api.CreateOptions) (*api.VEXDocument, error)
    MergeDocumentsFunc  func(context.Context, *api.MergeOptions) (*api.VEXDocument, error)
}

func (m *MockVEXClient) CreateStatement(ctx context.Context, opts *api.CreateOptions) (*api.VEXDocument, error) {
    if m.CreateStatementFunc != nil {
        return m.CreateStatementFunc(ctx, opts)
    }
    return &api.VEXDocument{ID: "test-doc"}, nil
}

func (m *MockVEXClient) MergeDocuments(ctx context.Context, opts *api.MergeOptions) (*api.VEXDocument, error) {
    if m.MergeDocumentsFunc != nil {
        return m.MergeDocumentsFunc(ctx, opts)
    }
    return &api.VEXDocument{ID: "merged-doc"}, nil
}

func (m *MockVEXClient) ValidateDocument(ctx context.Context, doc *api.VEXDocument) error {
    return nil
}

func (m *MockVEXClient) StreamMerge(ctx context.Context, opts *api.StreamMergeOptions) (<-chan *api.MergeResult, error) {
    ch := make(chan *api.MergeResult, 1)
    ch <- &api.MergeResult{Progress: 1.0, Status: "complete"}
    close(ch)
    return ch, nil
}

// Test utilities
func CreateTestRequest(t *testing.T, method string, params interface{}) *api.Request {
    t.Helper()
    
    return &api.Request{
        JSONRPC: "2.0",
        ID:      1,
        Method:  method,
        Params:  params,
    }
}

func AssertValidVEXDocument(t *testing.T, doc *api.VEXDocument) {
    t.Helper()
    
    if doc.ID == "" {
        t.Error("VEX document must have an ID")
    }
    
    if doc.Version == 0 {
        t.Error("VEX document must have a version")
    }
    
    if len(doc.Statements) == 0 {
        t.Error("VEX document should have at least one statement")
    }
}

func AssertMCPResponse(t *testing.T, resp *api.Response, expectError bool) {
    t.Helper()
    
    if resp.JSONRPC != "2.0" {
        t.Errorf("Expected JSONRPC 2.0, got %s", resp.JSONRPC)
    }
    
    if expectError && resp.Error == nil {
        t.Error("Expected error in response")
    }
    
    if !expectError && resp.Error != nil {
        t.Errorf("Unexpected error in response: %v", resp.Error)
    }
}
```

**Code Example - Testing Makefile:**
```makefile
# Makefile
.PHONY: test test-unit test-integration test-coverage

# Run all tests
test:
	go test ./...

# Run unit tests only
test-unit:
	go test -short ./...

# Run integration tests
test-integration:
	go test -run Integration ./...

# Run tests with coverage
test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Benchmark tests
benchmark:
	go test -bench=. ./...

# Clean test artifacts
clean-test:
	rm -f coverage.out coverage.html
```

**Deliverable**: Testing framework in `test/helpers/` and test configuration

---

## Phase 2 Deliverables

### 1. Core Architecture (`pkg/api/`)
- [ ] Complete interface definitions
- [ ] Type system for MCP and VEX operations
- [ ] Public API contracts

### 2. Configuration System (`internal/config/`)
- [ ] Multi-source configuration loading
- [ ] Environment variable support
- [ ] Configuration validation
- [ ] Default configuration values

### 3. Infrastructure (`internal/logging/`, `internal/errors/`)
- [ ] Structured logging with context support
- [ ] Error type system with categorization
- [ ] Request tracing capabilities

### 4. Testing Framework (`test/`)
- [ ] Mock implementations for testing
- [ ] Test helpers and utilities
- [ ] Coverage reporting setup
- [ ] Integration test structure

### 5. Project Structure
- [ ] Complete directory organization
- [ ] Go module dependencies
- [ ] Build configuration (Makefile)
- [ ] Documentation updates

## Success Criteria
- [ ] All interfaces compile and pass basic validation
- [ ] Configuration system works with environment variables and files
- [ ] Logging produces structured output at different levels
- [ ] Test framework supports unit and integration testing
- [ ] Project structure follows Go best practices
- [ ] Ready for MCP protocol implementation

## Dependencies
- **Input**: Phase 1 research findings and feasibility report
- **Output**: Solid foundation for MCP protocol implementation

## Risks & Mitigation
- **Risk**: Over-engineering interfaces before understanding requirements
  - **Mitigation**: Keep interfaces simple, iterate based on implementation needs
- **Risk**: Configuration complexity
  - **Mitigation**: Start with essential settings, expand gradually

## Time Estimate
**5 Story Points** ≈ 2-3 days of focused development

---
**Next**: [Phase 3: MCP Protocol Core](./phase3-mcp-core.md)
