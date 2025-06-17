#!/usr/bin/env node

import { spawn } from "child_process";
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { StreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/streamableHttp.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema
} from "@modelcontextprotocol/sdk/types.js";
import { tools, toolHandlers } from "./tools/index.js";

// Constants
const VEXCTL_CHECK_TIMEOUT = 5000; // 5 seconds timeout for vexctl check

// Check if vexctl is available
async function checkVexctlAvailability() {
  return new Promise((resolve) => {
    const vexctl = spawn("vexctl", ["version"], {
      stdio: ["ignore", "pipe", "pipe"]
    });
    
    let hasOutput = false;
    
    vexctl.stdout.on("data", () => {
      hasOutput = true;
    });
    
    vexctl.on("close", (code) => {
      resolve(hasOutput || code === 0);
    });
    
    vexctl.on("error", () => {
      resolve(false);
    });
    
    // Timeout after 5 seconds
    setTimeout(() => {
      vexctl.kill();
      resolve(false);
    }, VEXCTL_CHECK_TIMEOUT);
  });
}

// Create the server
const server = new Server(
  {
    name: "vexdoc-mcp-server",
    version: "1.0.0"
  },
  {
    capabilities: {
      tools: {}
    }
  }
);

// List available tools
server.setRequestHandler(ListToolsRequestSchema, async () => {
  return {
    tools: tools
  };
});

// Handle tool calls
server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;

  try {
    const handler = toolHandlers[name];
    if (!handler) {
      throw new Error(`Unknown tool: ${name}`);
    }
    
    return await handler(args);
  } catch (error) {
    return {
      content: [
        {
          type: "text",
          text: `Error: ${error.message}`
        }
      ],
      isError: true
    };
  }
});

// Error handling
server.onerror = (error) => {
  console.error("[MCP Error]", error);
};

process.on("SIGINT", async () => {
  await server.close();
  process.exit(0);
});

// Determine transport type and start the server
async function main() {
  // Check for vexctl availability before starting server
  console.error("🔍 Checking vexctl availability...");
  const vexctlAvailable = await checkVexctlAvailability();
  
  if (!vexctlAvailable) {
    console.error("❌ ERROR: vexctl command-line tool is not available");
    console.error("");
    console.error("The VEX Document MCP Server requires the 'vexctl' tool to be installed and available in your PATH.");
    console.error("");
    console.error("📦 Installation options:");
    console.error("  • Download from: https://github.com/openvex/vexctl/releases");
    console.error("  • Using Go: go install github.com/openvex/vexctl@latest");
    console.error("  • Using Homebrew: brew install vexctl");
    console.error("");
    console.error("After installation, ensure 'vexctl version' works from your terminal.");
    process.exit(1);
  }
  
  console.error("✅ vexctl is available");
  
  const args = process.argv.slice(2);
  const transportType = args[0] || "stdio";
  
  let transport;
  switch (transportType.toLowerCase()) {
  case "streaming":
  case "http": {
    const port = parseInt(args[1]) || 3000;
    transport = new StreamableHTTPServerTransport({
      port: port
    });
    console.error(`🚀 MCP Server running on HTTP transport at http://localhost:${port}`);
    break;
  }
      
  case "stdio":
  default:
    transport = new StdioServerTransport();
    console.error("🚀 MCP Server running on stdio transport");
    break;
  }
  
  await server.connect(transport);
}

main().catch((error) => {
  console.error("Failed to run server:", error);
  process.exit(1);
});
