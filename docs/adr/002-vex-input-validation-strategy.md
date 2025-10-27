# ADR 002: VEX Tool Input Validation Strategy

**Status**: Accepted  
**Date**: 2025-10-26  
**Deciders**: Development Team  
**Context**: Phase 4 - VEX Native Integration

## Context and Problem Statement

The Node.js VEX tools implement comprehensive input validation and sanitization to prevent security vulnerabilities. We need to decide how to implement equivalent validation in Go while maintaining security and following Go idioms.

## Decision Drivers

* **Security**: Must prevent injection attacks and malicious inputs
* **Consistency**: Match Node.js validation behavior
* **Go Idioms**: Follow Go error handling patterns
* **Performance**: Efficient validation without overhead
* **Maintainability**: Clear, testable validation code

## Node.js Validation Patterns

The Node.js implementation includes:

1. **Required field validation**
2. **Type checking** (string, array, object)
3. **Format validation** (CVE patterns, enums)
4. **Length limits** (prevent buffer overflows)
5. **Character blacklisting** (prevent command injection)
6. **Enum validation** (status, justification)
7. **Conditional validation** (status-specific requirements)

## Considered Options

### Option 1: Port Validation Logic Directly
**Approach**: Translate JavaScript validation to Go 1:1

**Pros**:
- Guaranteed parity with Node.js
- Clear mapping from original code
- Proven security properties

**Cons**:
- Not idiomatic Go (JavaScript patterns)
- May miss Go-specific validation opportunities
- Regex patterns may differ

### Option 2: Go-Idiomatic Validation with go-validator
**Approach**: Use struct tags with `go-validator` library

**Pros**:
- Declarative validation
- Standard Go pattern
- Built-in tag support

**Cons**:
- External dependency
- May not match all Node.js validation rules
- Less control over error messages

### Option 3: Custom Go Validation Package
**Approach**: Create internal validation package matching Node.js behavior

**Pros**:
- Full control over validation logic
- Go idioms with Node.js parity
- No external dependencies
- Clear error messages
- Testable validation functions

**Cons**:
- More upfront code to write
- Need to maintain validation package

## Decision Outcome

**Chosen option: Option 3 - Custom Go Validation Package**

Create `internal/vex/validation.go` that:
1. Ports all Node.js validation rules
2. Uses Go error handling patterns
3. Provides clear, specific error messages
4. Includes comprehensive unit tests

### Implementation Pattern

```go
// Validation functions return error for failures
func ValidateProduct(product string) error {
    if product == "" {
        return errors.New("product is required")
    }
    if len(product) > 1000 {
        return errors.New("product exceeds maximum length of 1000")
    }
    if containsDangerousChars(product) {
        return errors.New("product contains invalid characters")
    }
    return nil
}

// Composite validation
func ValidateCreateInput(input *CreateInput) error {
    if err := ValidateProduct(input.Product); err != nil {
        return err
    }
    if err := ValidateVulnerability(input.Vulnerability); err != nil {
        return err
    }
    // ... more validations
    return nil
}
```

## Validation Rules to Port

### 1. Required Fields
- Product, Vulnerability, Status (for create)
- Documents array (for merge, min 2 items)

### 2. String Length Limits
```
product:         1000 chars
vulnerability:   50 chars
impact_statement: 1000 chars
action_statement: 1000 chars
author:          200 chars
author_role:     200 chars
id:              500 chars
```

### 3. Format Validation
```go
// Vulnerability format: CVE-2023-1234, GHSA-xxxx-xxxx-xxxx
var vulnerabilityPattern = regexp.MustCompile(
    `^(CVE-\d{4}-\d+|GHSA-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}|[A-Z]+-\d+-\d+)$`)
```

### 4. Enum Validation
```go
var allowedStatuses = []string{
    "not_affected", "affected", "fixed", "under_investigation"}

var allowedJustifications = []string{
    "component_not_present",
    "vulnerable_code_not_present",
    "vulnerable_code_not_in_execute_path",
    "vulnerable_code_cannot_be_controlled_by_adversary",
    "inline_mitigations_already_exist"}
```

### 5. Dangerous Character Detection
```go
// Prevent injection attacks (even though we don't use subprocesses)
var dangerousChars = regexp.MustCompile(`[;&|` + "`" + `$(){}[\]<>'"\\]`)
```

### 6. Conditional Validation
- If `status == "not_affected"`, justification is required
- Validate document structure for merge operations
- Validate array bounds (2-20 documents for merge)

## Consequences

### Positive
- Full control over validation behavior
- Guaranteed parity with Node.js security
- No external dependencies
- Clear, testable validation code
- Go-idiomatic error handling
- Can optimize for performance

### Negative
- More code to write initially
- Need to maintain validation package
- Must update when Node.js validation changes

### Security Benefits
- Prevents malformed inputs
- Blocks injection attempts (defense in depth)
- Enforces business rules
- Provides clear error feedback
- Comprehensive test coverage

## Validation Testing Strategy

1. **Unit tests for each validation function**
2. **Table-driven tests** for enum validation
3. **Property-based tests** for regex patterns
4. **Fuzz testing** for dangerous character detection
5. **Integration tests** matching Node.js behavior

## References

- Node.js validation: `/src/tools/vex-schemas.js`
- Node.js create tool: `/src/tools/vex-create.js` (validateAndSanitizeInput)
- Node.js merge tool: `/src/tools/vex-merge.js` (validateAndSanitizeMergeInput)
- Go regexp package: https://pkg.go.dev/regexp
