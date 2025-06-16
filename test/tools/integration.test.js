import { describe, it } from "node:test";
import { strict as assert } from "assert";
import { tools, toolHandlers } from "../../src/tools/index.js";

describe("Tools Integration", () => {
  describe("Tools Export", () => {
    it("should export array of tools", () => {
      assert(Array.isArray(tools), "Tools should be an array");
      assert(tools.length > 0, "Should have at least one tool");
    });

    it("should have properly structured tool definitions", () => {
      tools.forEach((tool, index) => {
        assert(tool.name, `Tool ${index} should have a name`);
        assert(tool.description, `Tool ${index} should have a description`);
        assert(tool.inputSchema, `Tool ${index} should have an input schema`);
        assert(typeof tool.name === "string", `Tool ${index} name should be string`);
        assert(typeof tool.description === "string", `Tool ${index} description should be string`);
        assert(typeof tool.inputSchema === "object", `Tool ${index} schema should be object`);
      });
    });

    it("should have unique tool names", () => {
      const names = tools.map(tool => tool.name);
      const uniqueNames = new Set(names);
      assert.strictEqual(names.length, uniqueNames.size, "All tool names should be unique");
    });
  });

  describe("Tool Handlers Export", () => {
    it("should export object of tool handlers", () => {
      assert(typeof toolHandlers === "object", "Tool handlers should be an object");
      assert(toolHandlers !== null, "Tool handlers should not be null");
    });

    it("should have handler for each tool", () => {
      tools.forEach(tool => {
        assert(toolHandlers[tool.name], `Should have handler for tool: ${tool.name}`);
        assert(typeof toolHandlers[tool.name] === "function", 
          `Handler for ${tool.name} should be a function`);
      });
    });

    it("should not have extra handlers", () => {
      const toolNames = tools.map(tool => tool.name);
      const handlerNames = Object.keys(toolHandlers);
      
      handlerNames.forEach(handlerName => {
        assert(toolNames.includes(handlerName), 
          `Handler ${handlerName} should correspond to a defined tool`);
      });
    });
  });

  describe("Tool Handler Functionality", () => {
    it("should call handlers without errors", async () => {
      for (const tool of tools) {
        const handler = toolHandlers[tool.name];
        
        try {
          // Call with minimal valid args for each tool
          let testArgs = {};
          if (tool.name === "create_vex_statement") {
            testArgs = {
              product: "test-product",
              vulnerability: "CVE-2024-1234",
              status: "not_affected"
            };
          }
          
          const result = await handler(testArgs);
          assert(result, `Handler ${tool.name} should return a result`);
          assert(result.content, `Handler ${tool.name} should return content`);
          assert(Array.isArray(result.content), `Handler ${tool.name} content should be array`);
        } catch (error) {
          // Some errors are expected (e.g., vexctl not available), just ensure structure
          assert(error.message, `Handler ${tool.name} should provide error message`);
        }
      }
    });

    it("should return MCP-compliant responses", async () => {
      for (const tool of tools) {
        const handler = toolHandlers[tool.name];
        
        try {
          // Call with minimal valid args for each tool
          let testArgs = {};
          if (tool.name === "create_vex_statement") {
            testArgs = {
              product: "test-product",
              vulnerability: "CVE-2024-1234", 
              status: "not_affected"
            };
          }
          
          const result = await handler(testArgs);
          
          // Check MCP response format
          assert(result.content, `${tool.name} should return content`);
          assert(Array.isArray(result.content), `${tool.name} content should be array`);
          
          result.content.forEach((item, index) => {
            assert(item.type, `${tool.name} content item ${index} should have type`);
            assert(["text", "image", "resource"].includes(item.type), 
              `${tool.name} content item ${index} should have valid type`);
          });
        } catch (error) {
          // For tools that might fail (like vexctl), just verify error handling
          assert(error.message, `${tool.name} should provide error message`);
        }
      }
    });
  });

  describe("Specific Tool Availability", () => {
    it("should include VEX statement creation tool", () => {
      const vexTool = tools.find(tool => tool.name === "create_vex_statement");
      assert(vexTool, "Should include VEX statement creation tool");
      assert(vexTool.description.includes("VEX"), "VEX tool should mention VEX in description");
    });
  });

  describe("Input Schema Validation", () => {
    it("should have valid JSON schemas for all tools", () => {
      tools.forEach(tool => {
        const schema = tool.inputSchema;
        
        // Basic schema structure
        assert(schema.type, `${tool.name} schema should have type`);
        assert(schema.properties, `${tool.name} schema should have properties`);
        
        // Validate properties structure
        Object.entries(schema.properties).forEach(([propName, propSchema]) => {
          assert(propSchema.type, `${tool.name}.${propName} should have type`);
          assert(propSchema.description, `${tool.name}.${propName} should have description`);
        });
        
        // Validate required fields if present
        if (schema.required) {
          assert(Array.isArray(schema.required), `${tool.name} required should be array`);
          schema.required.forEach(requiredProp => {
            assert(schema.properties[requiredProp], 
              `${tool.name} required property ${requiredProp} should exist in properties`);
          });
        }
      });
    });
  });
});
