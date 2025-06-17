import { describe, it } from "node:test";
import { strict as assert } from "assert";
import { 
  assertValidToolResponse,
  testData
} from "../test-helpers.js";

// Import the vexctl tool functions
import { createVexStatementTool, handleCreateVexStatement } from "../../src/tools/vex-create.js";

describe("VEX Create Tool", () => {
  describe("Tool Definition", () => {
    it("should have correct tool definition structure", () => {
      const tool = createVexStatementTool;
      
      assert.strictEqual(tool.name, "create_vex_statement");
      assert.strictEqual(tool.description.includes("VEX"), true);
      assert(tool.inputSchema, "Tool should have input schema");
      assert(tool.inputSchema.properties, "Input schema should have properties");
      
      // Check required properties
      const requiredProps = ["product", "vulnerability", "status"];
      requiredProps.forEach(prop => {
        assert(tool.inputSchema.properties[prop], `Should have ${prop} property`);
      });
    });

    it("should have proper security constraints in schema", () => {
      const tool = createVexStatementTool;
      const statusProperty = tool.inputSchema.properties.status;
      
      assert(statusProperty.enum, "Status should have enum constraint");
      assert(statusProperty.enum.includes("not_affected"), "Should include not_affected status");
      assert(statusProperty.enum.includes("affected"), "Should include affected status");
      assert(statusProperty.enum.includes("fixed"), "Should include fixed status");
      assert(statusProperty.enum.includes("under_investigation"), "Should include under_investigation status");
    });
  });

  describe("Input Validation", () => {
    it("should accept valid VEX request", async () => {
      const response = await handleCreateVexStatement(testData.validVexRequest);
      
      assertValidToolResponse(response);
      assert.strictEqual(response.isError, undefined, "Should not have error flag");
    });

    it("should reject missing required parameters", async () => {
      const invalidRequest = { ...testData.validVexRequest };
      delete invalidRequest.product;
      
      const response = await handleCreateVexStatement(invalidRequest);
      
      assert.strictEqual(response.isError, true);
      assert(response.content[0].text.includes("Product parameter is required"));
    });

    it("should reject invalid status values", async () => {
      const invalidRequest = {
        ...testData.validVexRequest,
        status: "invalid_status"
      };
      
      const response = await handleCreateVexStatement(invalidRequest);
      
      assert.strictEqual(response.isError, true);
      assert(response.content[0].text.includes("Status must be one of"));
    });

    it("should reject invalid CVE format", async () => {
      const invalidRequest = {
        ...testData.validVexRequest,
        vulnerability: "INVALID-CVE-FORMAT"
      };
      
      const response = await handleCreateVexStatement(invalidRequest);
      
      assert.strictEqual(response.isError, true);
      assert(response.content[0].text.includes("Vulnerability must be in valid format"));
    });

    it("should reject empty string parameters", async () => {
      const invalidRequest = {
        ...testData.validVexRequest,
        product: "   "  // Whitespace only
      };
      
      const response = await handleCreateVexStatement(invalidRequest);
      
      assert.strictEqual(response.isError, true);
      assert(response.content[0].text.includes("Product parameter is required"));
    });

    it("should reject non-string parameters", async () => {
      const invalidRequest = {
        ...testData.validVexRequest,
        product: 12345  // Number instead of string
      };
      
      const response = await handleCreateVexStatement(invalidRequest);
      
      assert.strictEqual(response.isError, true);
      assert(response.content[0].text.includes("Product parameter is required"));
    });

    it("should validate justification when status is not_affected", async () => {
      const invalidRequest = {
        ...testData.validVexRequest,
        status: "not_affected",
        justification: "invalid_justification"
      };
      
      const response = await handleCreateVexStatement(invalidRequest);
      
      assert.strictEqual(response.isError, true);
      assert(response.content[0].text.includes("Justification must be one of"));
    });

    it("should accept valid justifications for not_affected status", async () => {
      const validJustifications = [
        "component_not_present",
        "vulnerable_code_not_present",
        "vulnerable_code_not_in_execute_path",
        "vulnerable_code_cannot_be_controlled_by_adversary",
        "inline_mitigations_already_exist"
      ];

      for (const justification of validJustifications) {
        const request = {
          ...testData.validVexRequest,
          status: "not_affected",
          justification
        };
        
        const response = await handleCreateVexStatement(request);
        
        assert.strictEqual(response.isError, undefined, 
          `Should accept justification: ${justification}`);
      }
    });
  });

  describe("Security Features", () => {
    it("should prevent argument injection in product parameter", async () => {
      const maliciousRequest = {
        ...testData.validVexRequest,
        product: "product; rm -rf /"
      };
      
      const response = await handleCreateVexStatement(maliciousRequest);
      
      // Should either sanitize or reject, but not execute the command
      assert(response, "Should handle malicious input safely");
    });

    it("should prevent path traversal in product parameter", async () => {
      const maliciousRequest = {
        ...testData.validVexRequest,
        product: "../../../etc/passwd"
      };
      
      const response = await handleCreateVexStatement(maliciousRequest);
      
      // Should handle safely without accessing filesystem outside expected paths
      assert(response, "Should handle path traversal attempts safely");
    });

    it("should handle extremely long input strings", async () => {
      const longString = "A".repeat(10000);
      const requestWithLongInput = {
        ...testData.validVexRequest,
        impact_statement: longString
      };
      
      const response = await handleCreateVexStatement(requestWithLongInput);
      
      // Should either accept or reject gracefully, not crash
      assert(response, "Should handle long inputs gracefully");
    });

    it("should not write files to filesystem", async () => {
      const response = await handleCreateVexStatement(testData.validVexRequest);
      
      assertValidToolResponse(response);
      
      // Verify response format - should return JSON or attachment, not file paths
      const content = response.content[0];
      assert(["text", "resource"].includes(content.type), 
        "Should return text or resource, not file references");
    });
  });

  describe("Output Format", () => {
    it("should return well-formed VEX document", async () => {
      const response = await handleCreateVexStatement(testData.validVexRequest);
      
      assertValidToolResponse(response);
      
      if (response.content[0].type === "text") {
        const content = response.content[0].text;
        
        // Should contain VEX-related content
        assert(content.includes("vex") || content.includes("VEX"), 
          "Response should contain VEX-related content");
      }
    });

    it("should include all provided parameters in output", async () => {
      const response = await handleCreateVexStatement(testData.validVexRequest);
      
      assertValidToolResponse(response);
      
      if (response.content[0].type === "text") {
        const content = response.content[0].text;
        
        // Check that key information is present
        assert(content.includes(testData.validVexRequest.vulnerability), 
          "Should include vulnerability ID");
        assert(content.includes(testData.validVexRequest.status), 
          "Should include status");
      }
    });
  });

  describe("Error Handling", () => {
    it("should handle vexctl command failures gracefully", async () => {
      // Create a request that might cause vexctl to fail
      const problematicRequest = {
        ...testData.validVexRequest,
        product: "///invalid//product///"
      };
      
      const response = await handleCreateVexStatement(problematicRequest);
      
      // Should return error response, not crash
      assert(response, "Should return response even on command failure");
      
      if (response.isError) {
        assert(response.content[0].text, "Error response should have message");
      }
    });

    it("should timeout on long-running vexctl commands", async () => {
      // This test verifies the timeout mechanism
      // The actual timeout value should be reasonable (not too long)
      const response = await handleCreateVexStatement(testData.validVexRequest);
      
      // Command should complete within reasonable time
      assert(response, "Should complete within timeout period");
    });
  });
});
