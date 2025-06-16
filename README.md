# VEX Document MCP Server

A Model Context Protocol (MCP) server for working with VEX (Vulnerability Exploitability eXchange) documents using the vexctl command-line tool.

## Features

- ✅ **Dual Transport Support** - stdio and Streamable HTTP
- ✅ **VEX Statement Creation** - Using vexctl command-line tool
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
# or 
npm run start:http
# or with custom port
node index.js http 8080
```

## Available Tools

### `create_vex_statement`
Creates a VEX statement using the vexctl command-line tool and returns the JSON content.

**Required Parameters:**
- `product` (string): Product identifier (PURL format recommended)
- `vulnerability` (string): Vulnerability ID (CVE-YYYY-NNNN format)
- `status` (enum): Impact status - `not_affected`, `affected`, `fixed`, `under_investigation`

**Optional Parameters:**
- `justification` (enum): Required for "not_affected" status
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
- ✅ **Injection Prevention** - No shell execution, argument array usage
- ✅ **Resource Limits** - Output size and execution time limits
- ✅ **Path Safety** - No file system access beyond vexctl execution
- ✅ **Error Sanitization** - Prevents information disclosure

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

## License

MIT
