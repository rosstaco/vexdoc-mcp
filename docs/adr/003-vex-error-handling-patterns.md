# ADR 003: VEX Tool Error Handling Patterns

**Status**: Accepted  
**Date**: 2025-10-26  
**Deciders**: Development Team  
**Context**: Phase 4 - VEX Native Integration

## Context and Problem Statement

VEX tools need consistent error handling that:
1. Matches Node.js tool behavior (error messages and format)
2. Follows Go error handling idioms
3. Provides clear, actionable error messages
4. Maintains security (no information disclosure)

## Decision Drivers

* **User Experience**: Clear error messages matching Node.js version
* **Go Idioms**: Proper error wrapping and handling
* **Security**: Prevent information disclosure in errors
* **Debugging**: Sufficient context for troubleshooting
* **Consistency**: Same error format across all tools

## Node.js Error Pattern

```javascript
// Node.js returns errors in content array
{
  content: [{
    type: "text",
    text: "Error: <message>"
  }],
  isError: true
}
```

## Considered Options

### Option 1: Direct Error Return
**Approach**: Return Go errors directly from tool functions

**Pros**:
- Simplest Go pattern
- Standard error handling

**Cons**:
- Doesn't match Node.js format
- MCP server needs to convert errors
- Inconsistent response structure

### Option 2: Result Wrapper Type
**Approach**: All tools return Result type with error field

**Pros**:
- Consistent response structure
- Matches Node.js pattern
- Clear success/error distinction

**Cons**:
- Deviates from Go conventions
- More boilerplate code

### Option 3: Dual Return Pattern
**Approach**: Return (result, error) - convert to MCP format at server layer

**Pros**:
- Idiomatic Go
- Clear separation of concerns
- Server handles MCP formatting
- Tools focus on business logic

**Cons**:
- Conversion logic in server
- Need to maintain formatting consistency

## Decision Outcome

**Chosen option: Option 3 - Dual Return Pattern**

### Pattern Structure

```go
// Tool implementation returns Go error
func CreateVexStatement(input *CreateInput) (*vex.VEX, error) {
    if err := validate(input); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    doc, err := vex.New()
    if err != nil {
        return nil, fmt.Errorf("failed to create document: %w", err)
    }
    
    return doc, nil
}

// Server converts to MCP ToolResult
func (t *VexCreateTool) Execute(ctx context.Context, args map[string]interface{}) 
    (*api.ToolResult, error) {
    
    input, err := parseCreateInput(args)
    if err != nil {
        return &api.ToolResult{
            Content: []api.Content{{
                Type: "text",
                Text: fmt.Sprintf("Error: %s", err.Error()),
            }},
            IsError: true,
        }, nil // Return nil error, error is in result
    }
    
    doc, err := CreateVexStatement(input)
    if err != nil {
        return &api.ToolResult{
            Content: []api.Content{{
                Type: "text", 
                Text: fmt.Sprintf("Error: %s", err.Error()),
            }},
            IsError: true,
        }, nil
    }
    
    // Success case
    return formatSuccessResult(doc), nil
}
```

## Error Categories

### 1. Validation Errors
**Cause**: Invalid input parameters  
**Example**: `"Error: product is required"`  
**Handling**: Return immediately with clear message

### 2. VEX Library Errors
**Cause**: VEX document creation/manipulation fails  
**Example**: `"Error: failed to create VEX document: invalid context"`  
**Handling**: Wrap with context using `fmt.Errorf`

### 3. Serialization Errors
**Cause**: JSON marshaling fails  
**Example**: `"Error: failed to format VEX document"`  
**Handling**: Should be rare, log for investigation

### 4. Unexpected Errors
**Cause**: System-level issues  
**Example**: `"Error: operation failed"`  
**Handling**: Generic message, log details securely

## Error Message Patterns

### Match Node.js Messages

| Scenario | Node.js Message | Go Message |
|----------|----------------|------------|
| Missing product | `"Error: Product parameter is required..."` | `"Error: product is required and must be a non-empty string"` |
| Invalid status | `"Error: Status must be one of: ..."` | `"Error: status must be one of: not_affected, affected, fixed, under_investigation"` |
| Invalid CVE | `"Error: Vulnerability must be in valid format..."` | `"Error: vulnerability must be in valid format (e.g., CVE-2023-1234)"` |
| Missing justification | `"Error: Justification is required when status is 'not_affected'"` | `"Error: justification is required when status is 'not_affected'"` |

### Error Message Guidelines

1. **Start with "Error: "** - Consistent prefix
2. **Be specific** - What field/validation failed
3. **Be actionable** - How to fix it
4. **Include examples** - Show valid formats
5. **No stack traces** - Keep user-facing messages clean
6. **No internal paths** - Prevent information disclosure

## Error Wrapping Strategy

```go
// Business logic errors - wrap with context
if err := validate(input); err != nil {
    return nil, fmt.Errorf("validation failed: %w", err)
}

// Library errors - wrap with operation context
doc, err := vex.New()
if err != nil {
    return nil, fmt.Errorf("failed to create VEX document: %w", err)
}

// Don't wrap user-facing validation errors
if input.Product == "" {
    return nil, errors.New("product is required") // Clear, direct
}
```

## Logging Strategy

```go
// Log errors for debugging, but don't expose to users
func (t *VexCreateTool) Execute(...) (*api.ToolResult, error) {
    doc, err := CreateVexStatement(input)
    if err != nil {
        // Log full error with context
        log.Printf("[ERROR] VEX create failed: %v", err)
        
        // Return sanitized error to user
        return &api.ToolResult{
            Content: []api.Content{{
                Type: "text",
                Text: fmt.Sprintf("Error: %s", sanitizeError(err)),
            }},
            IsError: true,
        }, nil
    }
}

func sanitizeError(err error) string {
    // Remove internal paths, stack traces, etc.
    msg := err.Error()
    // Keep only the user-relevant message
    return msg
}
```

## Security Considerations

### 1. No Information Disclosure
- Don't expose file paths
- Don't expose internal structure
- Don't expose system details
- Keep error messages user-focused

### 2. Consistent Error Format
- Same format prevents fingerprinting
- Generic fallback for unexpected errors
- No timing attacks via error messages

### 3. Error Message Testing
```go
func TestErrorMessages(t *testing.T) {
    tests := []struct {
        name     string
        input    CreateInput
        wantErr  string
    }{
        {
            name:    "missing product",
            input:   CreateInput{Vulnerability: "CVE-2023-1234"},
            wantErr: "Error: product is required",
        },
        // ... more cases matching Node.js
    }
}
```

## Consequences

### Positive
- Idiomatic Go error handling
- Clear separation of concerns
- Consistent error format with Node.js
- Secure error messages
- Easy to test error paths

### Negative
- Conversion logic in server layer
- Need to maintain message consistency
- More code for error formatting

### Testing Requirements
- Unit tests for all error conditions
- Integration tests comparing to Node.js errors
- Error message format validation
- Security testing (no information leaks)

## References

- Go error handling: https://go.dev/blog/error-handling-and-go
- Go errors package: https://pkg.go.dev/errors
- Node.js error examples: `/src/tools/vex-create.js`, `/src/tools/vex-merge.js`
- MCP ToolResult: `/go-implementation/pkg/api/types.go`
