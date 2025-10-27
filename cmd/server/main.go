package main

import (
	"log"
	"os"

	"github.com/rosstaco/vexdoc-mcp-go/internal/mcp"
	"github.com/rosstaco/vexdoc-mcp-go/internal/tools"
	"github.com/rosstaco/vexdoc-mcp-go/internal/vex"
)

func main() {
	// Create MCP server instance
	server := mcp.NewServer()

	// Create VEX client
	vexClient := vex.NewClient("vexdoc-mcp-server")

	// Register VEX tools
	createTool := tools.NewVEXCreateTool(vexClient)
	if err := server.RegisterTool(createTool); err != nil {
		log.Fatalf("Failed to register create tool: %v", err)
	}

	mergeTool := tools.NewVEXMergeTool(vexClient)
	if err := server.RegisterTool(mergeTool); err != nil {
		log.Fatalf("Failed to register merge tool: %v", err)
	}

	// Start server with stdio transport
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
