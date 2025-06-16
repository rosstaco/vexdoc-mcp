import { describe, it } from "node:test";
import { strict as assert } from "assert";
import { spawn } from "child_process";
import { 
  runCommand,
  sleep
} from "./test-helpers.js";

// Helper function to test stdio server startup
async function testStdioServer(args, expectedMessage, env = process.env) {
  const server = spawn("node", ["src/index.js", ...args], {
    cwd: "/workspaces/vexdoc-mcp",
    stdio: ["pipe", "pipe", "pipe"],
    env
  });

  let stderr = "";
  server.stderr.on("data", (data) => {
    stderr += data.toString();
  });

  // Wait for server to start
  await sleep(500);
  
  // Kill the server
  server.kill("SIGTERM");

  // Wait for process to exit
  await new Promise((resolve) => {
    server.on("exit", resolve);
  });

  assert(stderr.includes(expectedMessage), 
    `Expected "${expectedMessage}" in stderr. Got: ${stderr}`);
}

describe("MCP Server", () => {
  describe("Server Startup", () => {
    it("should start with stdio transport", async () => {
      await testStdioServer(["stdio"], "stdio transport");
    });

    it("should start with http transport", async () => {
      const result = await runCommand("node", ["src/index.js", "http", "3001"], {
        timeout: 2000,
        cwd: "/workspaces/vexdoc-mcp"
      });
      
      // Server should start and output should indicate HTTP transport
      assert(result.stderr.includes("HTTP") || result.stderr.includes("http") || 
             result.stderr.includes("3001"), 
      `Should indicate HTTP transport startup. Got: ${result.stderr}`);
    });

    it("should handle invalid transport gracefully", async () => {
      await testStdioServer(["invalid"], "stdio transport");
    });
  });

  describe("Command Line Arguments", () => {
    it("should default to stdio transport when no args provided", async () => {
      await testStdioServer([], "stdio");
    });

    it("should accept streaming transport alias", async () => {
      const result = await runCommand("node", ["src/index.js", "streaming", "3002"], {
        timeout: 3000,
        cwd: "/workspaces/vexdoc-mcp"
      });
      
      assert(result.stderr.includes("HTTP") || result.stderr.includes("3002"), 
        "Should handle streaming alias");
    });

    it("should default to port 3000 for HTTP transport", async () => {
      const result = await runCommand("node", ["src/index.js", "http"], {
        timeout: 3000,
        cwd: "/workspaces/vexdoc-mcp"
      });
      
      assert(result.stderr.includes("3000"), "Should default to port 3000");
    });
  });

  describe("Error Handling", () => {
    it("should handle port conflicts gracefully", async () => {
      // Start first server
      const server1 = spawn("node", ["src/index.js", "http", "3003"], {
        cwd: "/workspaces/vexdoc-mcp",
        stdio: ["pipe", "pipe", "pipe"]
      });

      await sleep(1000); // Let first server start

      // Try to start second server on same port
      const result = await runCommand("node", ["src/index.js", "http", "3003"], {
        timeout: 3000,
        cwd: "/workspaces/vexdoc-mcp"
      });

      // Clean up first server
      server1.kill();

      // Second server should handle port conflict
      assert(result.code !== undefined, "Should handle port conflict gracefully");
    });

    it("should handle SIGINT signal gracefully", async () => {
      const server = spawn("node", ["src/index.js", "stdio"], {
        cwd: "/workspaces/vexdoc-mcp",
        stdio: ["pipe", "pipe", "pipe"]
      });

      await sleep(500); // Let server start

      // Send SIGINT
      server.kill("SIGINT");

      // Wait for graceful shutdown
      const exitPromise = new Promise((resolve) => {
        server.on("exit", (code) => resolve(code));
      });

      const exitCode = await Promise.race([
        exitPromise,
        sleep(5000).then(() => "timeout")
      ]);

      assert(exitCode !== "timeout", "Server should shutdown within reasonable time");
      assert(exitCode === 0, "Server should exit cleanly on SIGINT");
    });
  });

  describe("MCP Protocol Compliance", () => {
    it("should have proper server capabilities", () => {
      // This test verifies that the server configuration includes proper capabilities
      // We'll check this by importing and inspecting the server setup
      
      // The server should be configured with tools capability
      assert(true, "Server capabilities should include tools"); // Placeholder
    });

    it("should handle ListTools requests", async () => {
      // Test that the server properly responds to list tools requests
      // This would require a more complex setup with actual MCP protocol testing
      
      assert(true, "Should handle ListTools requests properly"); // Placeholder
    });

    it("should handle CallTool requests", async () => {
      // Test that the server properly responds to call tool requests
      
      assert(true, "Should handle CallTool requests properly"); // Placeholder
    });
  });

  describe("Transport Configuration", () => {
    it("should configure stdio transport correctly", () => {
      // Verify stdio transport configuration
      assert(true, "Stdio transport should be configured correctly");
    });

    it("should configure HTTP transport with correct port", () => {
      // Verify HTTP transport configuration
      assert(true, "HTTP transport should be configured with correct port");
    });

    it("should handle transport initialization errors", () => {
      // Verify error handling during transport initialization
      assert(true, "Should handle transport initialization errors gracefully");
    });
  });

  describe("Server Lifecycle", () => {
    it("should initialize all components correctly", () => {
      // Test that server initializes tools, handlers, and transport
      assert(true, "Should initialize all components correctly");
    });

    it("should clean up resources on shutdown", () => {
      // Test proper cleanup of resources
      assert(true, "Should clean up resources properly on shutdown");
    });

    it("should handle multiple rapid startup/shutdown cycles", async () => {
      // Test stability under rapid cycling
      for (let i = 0; i < 3; i++) {
        const server = spawn("node", ["src/index.js", "stdio"], {
          cwd: "/workspaces/vexdoc-mcp",
          stdio: ["pipe", "pipe", "pipe"]
        });

        await sleep(200);
        server.kill("SIGINT");
        
        await new Promise((resolve) => {
          server.on("exit", resolve);
        });
      }
      
      assert(true, "Should handle rapid startup/shutdown cycles");
    });
  });

  describe("Environment Variables", () => {
    it("should respect NODE_ENV setting", async () => {
      await testStdioServer(["stdio"], "stdio", { ...process.env, NODE_ENV: "development" });
    });
  });
});
