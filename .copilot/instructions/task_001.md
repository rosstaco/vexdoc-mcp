# Task 001: Implement VEXctl Tool Integration

## Objective
Create an MCP tool that utilizes the `vexctl` Go command-line tool to create VEX (Vulnerability Exploitability eXchange) statements.

## Background
VEXctl is a command-line tool for working with VEX documents. We want to integrate it into our MCP server to allow clients to create VEX statements programmatically.

## Requirements

### Tool Specification
- **Tool Name**: `create_vex_statement`
- **Description**: Create a VEX statement using the vexctl command-line tool
- **Method**: Execute `vexctl` as a child process and return the result

### Parameters Needed
The tool should accept parameters for creating a VEX statement. Common vexctl parameters include:

**Required Parameters:**
- `vulnerability`: Vulnerability identifier (e.g., CVE-2023-1234)
- `product`: Product identifier (PURL format recommended)
- `status`: Vulnerability status (`not_affected`, `affected`, `fixed`, `under_investigation`)

**Optional Parameters:**
- `justification`: Reason for the status
- `impact_statement`: Description of impact
- `action_statement`: Actions taken or recommended
- `output_file`: Output file path (default: stdout)

### Implementation Notes
1. Validate input parameters before executing vexctl
2. Handle vexctl command execution errors gracefully
3. Return structured response with success/error status
4. Include the generated VEX statement in the response

### Example Usage
```bash
vexctl create --product "pkg:npm/example@1.0.0" --vulnerability "CVE-2023-1234" --status "not_affected" --justification "vulnerable_code_not_present"
```

## Success Criteria
- [ ] Tool can be called via MCP protocol
- [ ] Successfully executes vexctl command
- [ ] Returns VEX statement content
- [ ] Handles errors appropriately
- [ ] Validates input parameters

## Next Steps
1. Check if vexctl is available in the environment
2. Implement the tool in the MCP server
3. Test with various parameter combinations
4. Add proper error handling and validation

## Questions/Clarifications Needed
- What specific vexctl command format should we use?
- Are there any additional parameters we should support?
- Should we validate PURL format for products?
- Do we need to support batch creation of multiple statements?
