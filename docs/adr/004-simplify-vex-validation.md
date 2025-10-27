# ADR 004: Simplification of VEX Input Validation Using go-vex Library

**Status**: Accepted & Implemented (2025-10-26)  
**Date**: 2025-10-26  
**Deciders**: Development Team  
**Context**: Phase 4-5 Refinement

## Context and Problem Statement

We currently have extensive custom input validation in `internal/vex/validation.go` (190 lines) that validates VEX inputs before passing them to the go-vex library. After examining how `vexctl` (the reference implementation) handles validation, we need to determine:

1. Can we simplify or eliminate our custom validation?
2. Should we rely more on go-vex's built-in validation?
3. What validation should remain at the MCP layer vs domain layer?

## Decision Drivers

* **Simplicity**: Reduce duplicate validation logic
* **Maintainability**: Let the authoritative library handle domain validation
* **Defense in Depth**: Keep security-critical checks at boundaries
* **Consistency**: Follow patterns from vexctl reference implementation
* **Error Messages**: Provide clear, actionable feedback

## Research Findings

### How vexctl Handles Validation

From examining `openvex/vexctl` source code:

1. **Minimal Pre-validation** - vexctl validates at CLI layer, not domain layer:
   ```go
   // From vexctl/internal/cmd/options.go
   func (so *vexStatementOptions) Validate() error {
       if len(so.Products) == 0 {
           return errors.New("a minimum of one product id is needed")
       }
       if so.Vulnerability == "" {
           return errors.New("a vulnerability ID is required")
       }
       if so.Status == "" || !vex.Status(so.Status).Valid() {
           return fmt.Errorf("a valid impact status is required")
       }
       // ... conditional logic for status-specific requirements
   }
   ```

2. **Relies on go-vex Validation** - After basic checks, calls `statement.Validate()`:
   ```go
   // From vexctl/internal/cmd/add.go
   if err := statement.Validate(); err != nil {
       return fmt.Errorf("invalid statement: %w", err)
   }
   ```

3. **No String Length Checks** - vexctl does NOT validate:
   - Maximum string lengths
   - Dangerous character detection
   - Regex patterns for vulnerability formats
   - Product PURL format validation

4. **Status-Specific Logic** - Only validates conditional requirements:
   - `not_affected` requires justification
   - `affected` allows action statement
   - Justification only valid for `not_affected`

### What go-vex Validates

The `Statement.Validate()` method in go-vex handles:
- Required fields (vulnerability, products, status)
- Status enum validation via `Status.Valid()`
- Conditional requirements (justification for not_affected)
- Statement structure integrity

### Our Current Validation (190 lines)

We implement extensive checks that vexctl does NOT:
- ✅ String length limits (product: 1000, vuln: 50, etc.)
- ✅ Dangerous character detection (`[;&|`$(){}[]<>'"\\ ]`)
- ✅ Regex validation for CVE/GHSA format
- ✅ Enum validation for statuses and justifications
- ✅ Status-specific requirements

## Considered Options

### Option 1: Keep All Current Validation (Status Quo)
**Approach**: Maintain all 190 lines of custom validation

**Pros**:
- Maximum security (defense in depth)
- Consistent with Node.js implementation
- Clear error messages at MCP boundary
- Catches issues before go-vex processing

**Cons**:
- Duplicates go-vex validation logic
- More code to maintain
- Not following vexctl reference pattern
- May reject valid inputs go-vex would accept

### Option 2: Minimal Validation + go-vex (vexctl Pattern)
**Approach**: Only validate what's truly needed at MCP layer, rely on go-vex for domain validation

**Keep (Security/MCP Layer)**:
- Required field presence (product, vulnerability, status)
- Dangerous character detection (defense in depth)
- Basic string length limits (prevent DoS)
- Type checking (string vs array, etc.)

**Remove (Let go-vex Handle)**:
- Regex pattern matching for CVE/GHSA format
- Enum validation for statuses/justifications (use `Status.Valid()`)
- Status-specific conditional logic
- Detailed format validation

**Pros**:
- Simpler, more maintainable code
- Follows vexctl reference pattern
- Trusts authoritative library
- Automatically gets updates when go-vex changes
- Reduces code from 190 to ~60 lines

**Cons**:
- Less validation at MCP boundary
- Error messages might be less specific
- Slightly different from Node.js behavior

### Option 3: Hybrid Approach
**Approach**: Keep security checks, delegate domain checks to go-vex

**Keep**:
- String length limits (security)
- Dangerous character detection (security)
- Required field checks (early feedback)

**Delegate to go-vex**:
- CVE/GHSA format validation
- Status/justification enum validation
- Status-specific conditional logic

**Pros**:
- Balance between security and simplicity
- Maintains defense in depth for security
- Simpler domain validation
- ~100 lines of validation

**Cons**:
- Still some duplication
- Split responsibility

## Analysis: What Validation is Actually Needed?

### Security-Critical (Keep)
1. **String Length Limits**: Prevents DoS attacks
   ```go
   if len(product) > 1000 { return error }
   ```

2. **Dangerous Characters**: Defense in depth (even though we don't use subprocesses)
   ```go
   dangerousChars := regexp.MustCompile(`[;&|`$(){}[\]<>'"\\]`)
   ```

3. **Required Fields**: Early validation for better UX
   ```go
   if product == "" { return error }
   ```

### Domain Logic (Delegate to go-vex)
1. **Enum Validation**: Use `Status.Valid()` and `Justification.Valid()`
   ```go
   // Instead of our custom list, use:
   status, err := parseStatus(input.Status)
   if err != nil || !status.Valid() { return error }
   ```

2. **Format Validation**: Let go-vex validate CVE/GHSA patterns
   ```go
   // Remove our regex, let go-vex handle it
   ```

3. **Status-Specific Logic**: Use `Statement.Validate()`
   ```go
   // Remove our conditional checks, let statement.Validate() handle it
   ```

## Recommendation

**Chosen option: Option 2 - Minimal Validation + go-vex (vexctl Pattern)**

### Rationale

1. **Follow Reference Implementation**: vexctl is the authoritative CLI tool and doesn't do extensive pre-validation
2. **Trust the Library**: go-vex is the source of truth for VEX validation rules
3. **Reduce Maintenance**: Less code to maintain, automatic updates with go-vex
4. **Security Where It Matters**: Keep DoS prevention (length limits) and defense in depth (dangerous chars)
5. **Better Errors from Source**: go-vex error messages are authoritative

### Implementation Plan

#### Simplify `internal/vex/validation.go`

**Keep (Security Layer - ~60 lines)**:
```go
// Security limits
const (
    MaxProductLength         = 1000
    MaxVulnerabilityLength   = 50
    MaxStatementLength       = 1000
    MaxAuthorLength          = 200
    MaxMergeDocuments        = 20
)

// Security checks only
func ValidateStringLength(name, value string, max int) error
func ValidateDangerousChars(name, value string) error
func ValidateDocumentCount(count int) error
```

**Remove (Let go-vex handle)**:
- `AllowedStatuses` array (use `vex.Status.Valid()`)
- `AllowedJustifications` array (use `vex.Justification.Valid()`)
- `vulnerabilityPattern` regex
- `ValidateStatus()` (use `Status.Valid()`)
- `ValidateJustification()` (use `Justification.Valid()`)
- `ValidateStatusRequirements()` (use `Statement.Validate()`)

#### Update `internal/vex/client.go`

**Before**:
```go
func (c *Client) validateCreateInput(input *CreateInput) error {
    if err := ValidateProduct(input.Product); err != nil {
        return err
    }
    if err := ValidateVulnerability(input.Vulnerability); err != nil {
        return err
    }
    if err := ValidateStatus(input.Status); err != nil {
        return err
    }
    // ... 8 more validations
    if err := ValidateStatusRequirements(input.Status, input.Justification); err != nil {
        return err
    }
    return nil
}
```

**After**:
```go
func (c *Client) validateCreateInput(input *CreateInput) error {
    // Basic security checks only
    if err := ValidateStringLength("product", input.Product, MaxProductLength); err != nil {
        return err
    }
    if err := ValidateDangerousChars("product", input.Product); err != nil {
        return err
    }
    if err := ValidateStringLength("vulnerability", input.Vulnerability, MaxVulnerabilityLength); err != nil {
        return err
    }
    // Let go-vex handle domain validation via Statement.Validate()
    return nil
}
```

**Then in CreateStatement**:
```go
// Create statement
statement := vexlib.Statement{...}

// Let go-vex validate domain rules
if err := statement.Validate(); err != nil {
    return nil, fmt.Errorf("invalid VEX statement: %w", err)
}
```

## Consequences

### Positive
- **-130 lines of code** (190 → 60 lines)
- Follows vexctl reference pattern
- Simpler, more maintainable
- Automatic updates when go-vex changes
- Trust authoritative library for domain logic
- Clearer separation: security at boundary, domain in library

### Negative
- Different from Node.js implementation detail (but same API)
- Slightly less validation at MCP layer
- Error messages come from go-vex (may differ from our custom messages)
- Need to test that go-vex errors are user-friendly

### Neutral
- Still maintain string length limits (security)
- Still detect dangerous characters (defense in depth)
- Users get same end result (valid/invalid)

## Migration Steps

1. **Create new simplified validation.go**
2. **Update client.go to use minimal validation + Statement.Validate()**
3. **Add tests comparing error messages**
4. **Update documentation**
5. **Remove old validation functions**

## Testing Strategy

1. Test that go-vex rejects invalid status values
2. Test that go-vex enforces justification requirements
3. Test that our length limits still prevent DoS
4. Test that dangerous char detection still works
5. Compare error messages with Node.js (may differ in wording)

## Open Questions

1. **Are go-vex error messages user-friendly enough?**
   - Need to test actual error output
   - May need to wrap errors with additional context

2. **Should we keep CVE pattern validation?**
   - Leaning no - let go-vex be authoritative
   - Early rejection might be nice UX, but adds maintenance

3. **What about Node.js compatibility testing?**
   - Integration tests should verify same inputs accepted/rejected
   - Error message wording may differ

## References

- vexctl source: https://github.com/openvex/vexctl
- vexctl validation: `internal/cmd/options.go`, `internal/cmd/add.go`
- go-vex validation: `pkg/vex/vex.go` (`Statement.Validate()`)
- Our current validation: `internal/vex/validation.go`
- ADR 002: VEX Tool Input Validation Strategy

## Implementation

**Implemented**: 2025-10-26  
**Decision**: Option 2 - Balanced Approach (Simplified Security Boundaries)

See detailed implementation summary: `004-validation-simplification-implemented.md`

**Results**:
- Validation code reduced from 190 → 60 lines (68% reduction)
- All tests passing, server functioning correctly
- go-vex validation working as expected
- Security boundary checks maintained

