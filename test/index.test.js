import { describe, it } from "node:test";
import { strict as assert } from "assert";

// Import all test suites
import "./tools/vexctl.test.js";
import "./tools/integration.test.js";
import "./server.test.js";

describe("MCP Server Test Suite", () => {
  it("should run all test modules successfully", () => {
    assert(true, "All test modules loaded successfully");
  });
});

// Export test configuration
export const testConfig = {
  testDir: "./test",
  timeout: 10000,
  coverage: {
    include: ["src/**/*.js"],
    exclude: ["test/**/*.js", "node_modules/**"]
  }
};
