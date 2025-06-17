# VEX Document MCP Server

A Model Context Protocol (MCP) server for working with VEX (Vulnerability Exploitability eXchange) documents using the vexctl command-line tool. Designed for integration with AI-powered vulnerability assessment workflows.

## Features

- ✅ **Dual Transport Support** - stdio and Streamable HTTP
- ✅ **VEX Statement Creation** - Generate individual VEX documents
- ✅ **VEX Document Merging** - Consolidate multiple VEX documents 
- ✅ **AI Workflow Integration** - Perfect for Trivy + AI code analysis pipelines
- ✅ **Security Hardened** - Input validation and injection prevention
- ✅ **MCP Compliant** - Returns JSON content directly (no file writing)

## Installation

```bash
npm install
```

## Usage

### stdio Transport (default)
```bash
npm start
# or
npm run start:stdio
```

### Streamable HTTP Transport
```bash
npm run start:streaming
# or with custom port
node index.js streaming 8080
```

## Available Tools

### `create_vex_statement`
Creates a VEX statement for a specific vulnerability using the vexctl command-line tool.

**Required Parameters:**
- `product` (string): Product identifier (PURL format recommended)
- `vulnerability` (string): Vulnerability ID (CVE-YYYY-NNNN format)
- `status` (enum): Impact status - `not_affected`, `affected`, `fixed`, `under_investigation`

**Optional Parameters:**
- `justification` (enum): Required for "not_affected" status
  - `component_not_present`
  - `vulnerable_code_not_present`
  - `vulnerable_code_not_in_execute_path`
  - `vulnerable_code_cannot_be_controlled_by_adversary`
  - `inline_mitigations_already_exist`
- `impact_statement` (string): Explanation text
- `action_statement` (string): Action description  
- `author` (string): Document author

**Example:**
```json
{
  "product": "pkg:apk/wolfi/git@2.39.0-r1?arch=x86_64",
  "vulnerability": "CVE-2023-1234", 
  "status": "not_affected",
  "justification": "vulnerable_code_not_present"
}
```

### `merge_vex_documents`
Merge and consolidate multiple VEX documents into a unified security assessment report. Perfect for combining results from multiple vulnerability scans or different teams.

**Required Parameters:**
- `documents` (array): Array of VEX document objects (2-20 documents)

**Optional Parameters:**
- `author` (string): Author of the merged document
- `author_role` (string): Role of the author
- `id` (string): Unique identifier for the merged document
- `product_filter` (array): Filter by specific product identifiers
- `vulnerability_filter` (array): Filter by specific vulnerability IDs

**Example:**
```json
{
  "documents": [
    {
      "@context": "https://openvex.dev/ns/v0.2.0",
      "@id": "doc1",
      "statements": [...]
    },
    {
      "@context": "https://openvex.dev/ns/v0.2.0", 
      "@id": "doc2",
      "statements": [...]
    }
  ],
  "author": "Security Team",
  "author_role": "Security Engineer"
}
```

## Use Cases

### AI-Powered Vulnerability Assessment Pipeline
1. **Trivy Scan**: Identify CVEs in dependencies and containers
2. **AI Code Analysis**: Determine actual exploitability 
3. **VEX Generation**: Create standardized assessments using `create_vex_statement`
4. **Report Consolidation**: Merge findings using `merge_vex_documents`

### Enterprise Security Reporting
- Consolidate VEX documents from multiple repositories
- Create executive-level vulnerability exposure summaries
- Generate compliance documentation with proper audit trails

## Transport Types

### stdio Transport
- **Use Case**: Direct integration with MCP clients
- **Connection**: Standard input/output streams
- **Best For**: Command-line tools, desktop applications

### Streamable HTTP Transport  
- **Use Case**: Web-based integrations, API access
- **Connection**: HTTP server on specified port
- **Best For**: Web applications, browser extensions, remote access
- **Default Port**: 3000

## MCP Client Configuration

### Local Development (Recommended)
For development, use the local Node.js version:
```json
{
  "servers": {
    "vexdoc": {
      "command": "node",
      "args": ["/path/to/vexdoc-mcp/index.js", "stdio"],
      "env": {}
    }
  }
}
```

### Docker Configuration

First, build the Docker image:
```bash
docker build -t vexdoc-mcp .
```

#### Option 1: Docker with stdio transport
```json
{
  "servers": {
    "vexdoc-docker": {
      "command": "docker",
      "args": [
        "run", 
        "--rm", 
        "-i",
        "--init",
        "-e", "MCP_TRANSPORT=stdio",
        "vexdoc-mcp"
      ],
      "env": {}
    }
  }
}
```

#### Option 2: Pre-built image from registry (if published)
```json
{
  "servers": {
    "vexdoc": {
      "command": "docker",
      "args": [
        "run", 
        "--rm", 
        "-i",
        "--init",
        "-e", "MCP_TRANSPORT=stdio",
        "ghcr.io/your-org/vexdoc-mcp:latest"
      ],
      "env": {}
    }
  }
}
```

### Docker Arguments Explained

**Essential Docker arguments for MCP stdio transport:**
- `run` - Create and start container
- `--rm` - Remove container when it exits (cleanup)
- `-i` - Keep STDIN open (required for stdio transport)
- `--init` - Proper signal handling (recommended)
- `-e MCP_TRANSPORT=stdio` - Set transport mode
- `vexdoc-mcp` - Your built image name

**Optional arguments:**
- `--name container-name` - Give container a name
- `-v /host/path:/container/path` - Mount volumes if needed

## Security Features

- ✅ **Input Validation** - Strict parameter validation with allowlists
- ✅ **Injection Prevention** - Command injection protection for all inputs
- ✅ **Resource Limits** - Maximum 20 documents per merge, timeout protection
- ✅ **Path Safety** - Secure temporary file handling with automatic cleanup
- ✅ **Error Sanitization** - Prevents information disclosure
- ✅ **Schema Validation** - JSON schema validation for all tool inputs

## Docker Usage

### Quick Start with Docker Compose

```bash
# Start HTTP server (default)
docker-compose up -d

# Start development server with hot reload
docker-compose --profile dev up -d

# Start stdio server (for testing)
docker-compose --profile stdio up

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Manual Docker Commands

```bash
# Build the image
docker build -t vexdoc-mcp .

# Run HTTP server
docker run -d -p 3000:3000 -e MCP_TRANSPORT=http vexdoc-mcp

# Run stdio server
docker run -it -e MCP_TRANSPORT=stdio vexdoc-mcp

# Run with custom port
docker run -d -p 8080:8080 -e MCP_TRANSPORT=http -e MCP_PORT=8080 vexdoc-mcp
```

## Requirements

- Node.js 18.0.0 or higher
- vexctl command-line tool installed and available in PATH
- npm

## Development & Testing

### Running Tests
```bash
# Run all tests
npm test

# Run tests with coverage
npm run test:coverage

# Run linting
npm run lint
```

### Test Coverage
- **144 tests** covering all functionality
- **100% coverage** of critical paths
- Security validation and injection prevention tests
- Integration tests with vexctl command-line tool
- Error handling and edge case validation

### Project Structure
```
src/tools/
├── index.js           # Main tools export
├── vex-create.js      # VEX statement creation tool
├── vex-merge.js       # VEX document merging tool  
└── vex-schemas.js     # Shared schemas and validation
```

## License

MIT
