# Testing Documentation

## Overview

This MCP server includes comprehensive unit testing to ensure reliability, security, and proper functionality. The test suite covers:

- **Tool Functionality**: Each tool (echo, vexctl) is thoroughly tested
- **Input Validation**: Security and data validation testing
- **MCP Protocol Compliance**: Ensures proper MCP response formats
- **Server Lifecycle**: Startup, shutdown, and error handling
- **Integration Testing**: Tools working together

## Running Tests

### Quick Start
```bash
# Run all tests
npm test

# Run tests with coverage
npm run test:coverage

# Run tests in watch mode (for development)
npm run test:watch

# Run linting
npm run lint

# Fix linting issues automatically
npm run lint:fix
```

### Test Organization

```
test/
├── test-helpers.js         # Shared testing utilities
├── server.test.js          # Main MCP server tests
├── tools/
│   ├── echo.test.js        # Echo tool tests
│   ├── vexctl.test.js      # VEX tool tests
│   └── integration.test.js # Tools integration tests
└── index.test.js           # Test suite entry point
```

## Test Categories

### 1. Tool Tests (`test/tools/`)

#### Echo Tool (`echo.test.js`)
- Tool definition validation
- Basic echo functionality
- Format options (uppercase, lowercase, reverse)
- Repeat functionality
- Input validation and error handling
- Edge cases (unicode, long messages, newlines)

#### VEX Tool (`vexctl.test.js`)
- Tool definition and schema validation
- Input validation (required fields, format validation)
- Security features (injection prevention, path traversal)
- VEX document generation and format
- Error handling and timeouts
- Status-specific logic (justifications, action statements)

#### Integration Tests (`integration.test.js`)
- Tool export validation
- Handler consistency
- MCP response compliance
- Schema validation across all tools

### 2. Server Tests (`server.test.js`)

- **Startup Tests**: Different transport modes (stdio, HTTP)
- **Command Line Arguments**: Argument parsing and defaults
- **Error Handling**: Port conflicts, invalid inputs, signal handling
- **MCP Protocol**: Capability advertisement, request handling
- **Lifecycle Management**: Initialization, shutdown, resource cleanup

### 3. Security Testing

Throughout all test files, security aspects are validated:
- Input sanitization and validation
- Command injection prevention
- Path traversal protection
- Resource abuse prevention (timeouts, size limits)
- Error message sanitization

## Test Utilities (`test-helpers.js`)

### Mock Objects
- `MockTransport`: Simulates MCP transport for testing
- `createMockRequest()`: Creates properly formatted MCP requests
- `createToolCallRequest()`: Creates tool call requests

### Validation Helpers
- `assertValidMCPResponse()`: Validates MCP response format
- `assertValidToolResponse()`: Validates tool response structure
- `assertErrorMessage()`: Validates error message patterns

### Execution Helpers
- `runCommand()`: Safely executes commands with timeouts
- `sleep()`: Async delay utility
- Test data factories for common test scenarios

## Coverage and Quality

### Coverage Targets
- **Lines**: 80% minimum
- **Functions**: 80% minimum
- **Branches**: 80% minimum
- **Statements**: 80% minimum

### Code Quality
- ESLint configuration ensures consistent code style
- Security-focused linting rules
- Complexity limits to maintain maintainability

## Continuous Integration

The test suite is designed to run in various environments:
- **Local Development**: Fast feedback during development
- **Docker Containers**: Consistent environment testing
- **CI/CD Pipelines**: Automated testing and quality gates

## Known Test Patterns

### 1. Error Testing Pattern
```javascript
it('should handle invalid input', async () => {
  const response = await handler(invalidInput);
  assert.strictEqual(response.isError, true);
  assert(response.content[0].text.includes('expected error message'));
});
```

### 2. Security Testing Pattern
```javascript
it('should prevent injection attacks', async () => {
  const maliciousInput = { param: 'value; rm -rf /' };
  const response = await handler(maliciousInput);
  // Should either sanitize or reject safely
  assert(response);
});
```

### 3. MCP Compliance Pattern
```javascript
it('should return MCP-compliant response', async () => {
  const response = await handler(validInput);
  assertValidToolResponse(response);
  assert(response.content[0].type === 'text');
});
```

## Debugging Tests

### Running Individual Test Files
```bash
node --test test/tools/echo.test.js
node --test test/tools/vexctl.test.js
node --test test/server.test.js
```

### Verbose Output
```bash
node --test --test-reporter=spec
```

### Coverage Reports
```bash
npm run test:coverage
# Open coverage/index.html in browser
```

## Contributing to Tests

### Adding New Tests
1. Follow existing patterns and naming conventions
2. Include both positive and negative test cases
3. Test security aspects and edge cases
4. Update this documentation for new test categories

### Test Data
- Use the test data factories in `test-helpers.js`
- Create realistic but safe test data
- Avoid hardcoded values that might become stale

### Performance Considerations
- Use timeouts appropriately for external commands
- Mock external dependencies when possible
- Keep test execution time reasonable (< 30 seconds total)

This comprehensive test suite ensures the MCP server is reliable, secure, and maintainable across different environments and use cases.
