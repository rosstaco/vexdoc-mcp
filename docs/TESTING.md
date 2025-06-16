# Testing Guide

## Unit Testing for MCP Server

This project includes comprehensive unit testing using Node.js built-in test runner.

## Running Tests

```bash
# Run all tests
npm test

# Run with coverage
npm run test:coverage

# Run in watch mode  
npm run test:watch

# Run specific test file
npm test -- test/tools/vexctl.test.js
```

## Test Structure

- `test/tools/vexctl.test.js` - VEX tool functionality tests
- `test/tools/integration.test.js` - Tools integration tests  
- `test/server.test.js` - MCP server tests
- `test/test-helpers.js` - Testing utilities and test data

## Coverage

Minimum coverage requirements:
- Branches: 80%
- Functions: 80%
- Lines: 80%
- Statements: 80%

## Key Test Areas

### VEX Tool Tests
- Input validation and sanitization
- Security (injection prevention, length limits)
- vexctl command integration
- Error handling
- Output format validation

### Server Tests  
- Transport configuration (stdio, HTTP)
- Startup and shutdown procedures
- Command line argument handling
- Error scenarios

### Integration Tests
- Tool registry validation
- MCP protocol compliance
- Response format validation

## Security Testing

All tests include security validation:
- Command injection prevention
- Path traversal protection
- Input sanitization
- Resource limit enforcement

## Running Individual Test Categories

```bash
# VEX tool tests only
npm test -- test/tools/vexctl.test.js

# Server tests only  
npm test -- test/server.test.js

# Integration tests only
npm test -- test/tools/integration.test.js
```
