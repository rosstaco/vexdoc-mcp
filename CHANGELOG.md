# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- GitHub Actions CI/CD pipeline with automated releases
- GoReleaser configuration for multi-platform builds
- Semantic versioning support with pre-release tags
- Build-time version injection via ldflags
- Automated binary builds for Linux, macOS, and Windows (amd64 & arm64)

## [0.1.0] - 2024-10-27

### Added
- Initial Go implementation of VexDoc MCP server
- Native VEX library integration using go-vex v0.2.7
- MCP protocol implementation with JSON-RPC 2.0
- Standard I/O transport layer
- `create_vex_statement` tool with full validation
- `merge_vex_documents` tool with filtering support
- Comprehensive unit tests (87.4% VEX layer, 94.4% tools layer coverage)
- Security features: input validation, DoS prevention, injection defense
- Build system using Just with 28+ recipes
- Architecture Decision Records documenting design choices

### Security
- Input validation preventing injection attacks
- String length limits for DoS prevention
- Dangerous character detection
- Domain validation via go-vex library
- Safe error messages without information disclosure

[Unreleased]: https://github.com/rosstaco/vexdoc-mcp/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/rosstaco/vexdoc-mcp/releases/tag/v0.1.0
