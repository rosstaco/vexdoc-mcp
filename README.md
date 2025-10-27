# VexDoc MCP Server

A Model Context Protocol (MCP) server for VEX (Vulnerability Exploitability eXchange) document operations, written in Go.

## Overview

This project provides a high-performance MCP server that enables AI assistants to create, merge, and manipulate VEX documents through standardized tool interfaces.

## Architecture

```
vexdoc-mcp/
├── cmd/
│   └── server/          # Main server executable
├── internal/
│   ├── mcp/            # MCP protocol implementation
│   ├── tools/          # VEX tool implementations
│   └── vex/            # VEX library integration
├── pkg/
│   └── api/            # Public API definitions
├── docs/               # Documentation and ADRs
└── research/           # Research materials
```

## Development Status

**Phase 1: Research & Discovery** ✅ COMPLETE
- VEX Go library analysis completed
- MCP protocol research completed  
- Development environment established

**Phase 2: Project Foundation** ✅ COMPLETE
- ✅ Core MCP server implementation with JSON-RPC 2.0
- ✅ Stdio transport layer for MCP communication
- ✅ Type definitions for MCP protocol
- ✅ Tool registry and interface system
- ✅ Comprehensive test suite (38.9% MCP coverage)
- ✅ Build system using Just (justfile with 28 recipes)
- ✅ Project structure following Go best practices

**Phase 4: VEX Native Integration** ✅ COMPLETE
- ✅ Native VEX library integration (go-vex v0.2.7)
- ✅ Simplified validation following vexctl patterns (60 lines, 68% reduction)
- ✅ VEX client with create and merge operations
- ✅ go-vex domain validation integration
- ✅ Security boundary checks (DoS prevention, injection defense)
- ✅ Architecture Decision Records (4 ADRs documenting design choices)

**Phase 5: Tool Implementation** ✅ COMPLETE
- ✅ `create_vex_statement` tool with full validation (94.4% test coverage)
- ✅ `merge_vex_documents` tool with filtering support (94.4% test coverage)
- ✅ Error handling matching Node.js patterns
- ✅ JSON schema definitions for MCP
- ✅ Feature parity with Node.js implementation

**Phase 7: Testing** ✅ COMPLETE
- ✅ Comprehensive unit tests (33 passing tests, 0 failures)
- ✅ Validation layer tests (87.4% coverage)
- ✅ VEX client tests (87.4% coverage)
- ✅ Tool layer tests (94.4% coverage)
- ✅ Table-driven test approach for maintainability
- ✅ Success scenarios, error paths, and edge cases covered

**Phase 3: MCP Protocol Core** 🔄 DEFERRED
- HTTP transport implementation (future consideration if needed)
- Streaming support enhancements (future consideration)

**Phases 6: Performance** ⏭️ NEXT
- Performance optimization and benchmarking
- Load testing and profiling
- Memory usage optimization

**Phase 8: Deployment** ⏭️ FUTURE
- Integration testing with real MCP clients
- Production deployment configuration
- Documentation and examples

## Getting Started

### Prerequisites
- Go 1.21 or later
- Git

### Installation
```bash
git clone https://github.com/rosstaco/vexdoc-mcp.git
cd vexdoc-mcp
go mod tidy
```

### Build
```bash
# Using Just (recommended)
just build

# Or using Go directly
go build -o vexdoc-mcp-server ./cmd/server
```

### Run
```bash
./vexdoc-mcp-server
```

### Development Commands
```bash
just test          # Run tests
just coverage      # Generate coverage report
just lint          # Run linters
just fmt           # Format code
just ci            # Run all CI checks
just stats         # Show code statistics
just help          # Show all available commands
```

## Features (Implemented)

### VEX Operations
- ✅ VEX document creation (native Go library)
- ✅ VEX document merging (native Go library)
- ✅ Product and vulnerability filtering
- ✅ Input validation and sanitization
- ✅ OpenVEX v0.2.7 compliance

### MCP Integration
- ✅ JSON-RPC 2.0 protocol support
- ✅ Standard I/O transport
- ✅ Tool registration and discovery
- ✅ Comprehensive error handling and validation
- ✅ JSON schema support for tool inputs

### Security
- ✅ Input validation preventing injection attacks
- ✅ String length limits (DoS prevention)
- ✅ Dangerous character detection (defense in depth)
- ✅ Domain validation via go-vex library (single source of truth)
- ✅ Safe error messages (no information disclosure)

### Quality Assurance
- ✅ 33 unit tests covering all major functionality
- ✅ 87.4% test coverage for VEX layer
- ✅ 94.4% test coverage for tools layer
- ✅ Table-driven tests for maintainability
- ✅ CI/CD checks: fmt, vet, test, build

## Performance Targets
- **Startup time**: <50ms (currently ~10ms)
- **Memory usage**: <10MB baseline (currently ~3MB)
- **JSON processing**: Expected 2-3x faster than Node.js
- **Concurrent requests**: Native goroutine support
- **Binary size**: 2.9MB (optimized build)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.
