#!/usr/bin/env node

import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { StreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/streamableHttp.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema
} from "@modelcontextprotocol/sdk/types.js";
import { tools, toolHandlers } from "./tools/index.js";

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
    console.error(`MCP Server running on HTTP transport at http://localhost:${port}`);
    break;
  }
      
  case "stdio":
  default:
    transport = new StdioServerTransport();
    console.error("MCP Server running on stdio transport");
    break;
  }
  
  await server.connect(transport);
}

main().catch((error) => {
  console.error("Failed to run server:", error);
  process.exit(1);
});
