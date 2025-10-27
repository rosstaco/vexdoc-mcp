# ADR 001: VEX Tool Implementation Strategy

**Status**: Accepted  
**Date**: 2025-10-26  
**Deciders**: Development Team  
**Context**: Phase 4 - VEX Native Integration

## Context and Problem Statement

We need to implement VEX (Vulnerability Exploitability eXchange) tools for the MCP server in Go. The Node.js version uses subprocess execution of `vexctl` CLI commands. We must decide whether to:

1. Continue using subprocess execution of `vexctl` (like Node.js)
2. Use the Go VEX library directly (native integration)
3. Use a hybrid approach

## Decision Drivers

* **Performance**: Eliminate subprocess overhead
* **Reliability**: Reduce failure points from process spawning
* **Maintainability**: Simpler error handling and debugging
* **Security**: Reduce command injection attack surface
* **Feature Parity**: Must match Node.js implementation functionality
* **Migration Goal**: Phase 1 research validated native integration feasibility

## Considered Options

### Option 1: Subprocess Execution (Node.js Pattern)
**Pros**:
- Direct port of existing logic
- Proven working implementation
- Uses battle-tested vexctl CLI
- Easy validation against Node.js version

**Cons**:
- Subprocess overhead (~50-100ms per call)
- Complex error handling across process boundaries
- Security concerns (command injection, though mitigated)
- JSON serialization/deserialization overhead
- Defeats primary migration goal

### Option 2: Native Go Library Integration
**Pros**:
- Direct use of `github.com/openvex/go-vex` library
- 2-3x performance improvement (from Phase 1 research)
- Type-safe operations
- Simplified error handling
- Achieves migration goals
- No external process dependencies

**Cons**:
- Different code path than Node.js
- Need to ensure feature parity
- More upfront implementation effort

### Option 3: Hybrid Approach
**Pros**:
- Native for common operations
- Subprocess fallback for edge cases

**Cons**:
- Increased complexity
- Inconsistent behavior
- Defeats purpose of migration

## Decision Outcome

**Chosen option: Option 2 - Native Go Library Integration**

### Rationale

1. **Achieves Migration Goals**: Phase 1 research specifically validated that native integration would provide significant performance benefits
2. **Phase 1 Validation**: Research confirmed `github.com/openvex/go-vex` v0.2.5 provides equivalent or superior functionality
3. **Performance**: Expected 2-3x speed improvement with 50-70% memory reduction
4. **Security**: Eliminates command injection attack surface entirely
5. **Simplicity**: No process management, cleaner error handling
6. **Maintainability**: Single language stack, easier debugging

### Implementation Strategy

1. **Use go-vex library directly** for all VEX operations
2. **Follow Node.js tool patterns** for:
   - Input validation and sanitization
   - Error handling and messages
   - Response formatting
   - Security constraints (max lengths, character validation)
3. **Maintain API compatibility** with Node.js version
4. **Reuse validation logic** from Node.js (port to Go)

## Consequences

### Positive
- Faster tool execution (2-3x improvement)
- Lower memory usage (50-70% reduction)
- Simpler codebase (no subprocess management)
- Better error messages (native Go errors)
- Type safety at compile time
- Single binary deployment

### Negative
- Implementation differs from Node.js (but same API)
- Need comprehensive testing for feature parity
- Cannot use vexctl CLI directly for debugging

### Neutral
- Different internal implementation but same external behavior
- Need to document differences for maintenance

## Validation

Success criteria:
- ✅ All Node.js tool features implemented
- ✅ Same input schemas and validation rules
- ✅ Same output format (JSON VEX documents)
- ✅ Same error messages and handling
- ✅ Performance benchmarks show 2x+ improvement
- ✅ Security validation passes (no injection vulnerabilities)

## Related Decisions

- ADR 002: VEX Tool Input Validation Strategy
- ADR 003: VEX Tool Error Handling Patterns
- Phase 1 Research: VEX Library Analysis
- Phase 1 Research: MCP Protocol Implementation

## References

- Node.js implementation: `/src/tools/vex-create.js`, `/src/tools/vex-merge.js`
- Go VEX library: `github.com/openvex/go-vex`
- Phase 1 Research: `/research/research-vex-apis.md`
- Migration Plan: `/docs/GO_MIGRATION_PLAN.md`
