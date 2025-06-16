import { strict as assert } from "assert";
import { spawn } from "child_process";
import { EventEmitter } from "events";

/**
 * Test helper utilities for MCP Server testing
 */

/**
 * Mock transport for testing MCP server functionality
 */
export class MockTransport extends EventEmitter {
  constructor() {
    super();
    this.messages = [];
    this.connected = false;
  }

  connect() {
    this.connected = true;
    return Promise.resolve();
  }

  close() {
    this.connected = false;
    return Promise.resolve();
  }

  send(message) {
    this.messages.push(message);
    this.emit("message", message);
  }

  getMessages() {
    return [...this.messages];
  }

  clearMessages() {
    this.messages = [];
  }
}

/**
 * Create a mock MCP request for testing
 */
export function createMockRequest(method, params = {}, id = "test-123") {
  return {
    jsonrpc: "2.0",
    id,
    method,
    params
  };
}

/**
 * Create a mock tool call request
 */
export function createToolCallRequest(toolName, args = {}, id = "test-tool-123") {
  return createMockRequest("tools/call", {
    name: toolName,
    arguments: args
  }, id);
}

/**
 * Assert that a response has the expected structure
 */
export function assertValidMCPResponse(response, expectedId = null) {
  assert(response, "Response should exist");
  assert.strictEqual(response.jsonrpc, "2.0", "Response should have jsonrpc 2.0");
  
  if (expectedId) {
    assert.strictEqual(response.id, expectedId, "Response should have correct ID");
  }
  
  assert(response.result || response.error, "Response should have result or error");
}

/**
 * Assert that a tool response has the expected structure
 */
export function assertValidToolResponse(response) {
  assert(response, "Tool response should exist");
  assert(response.content, "Tool response should have content");
  assert(Array.isArray(response.content), "Tool response content should be an array");
  
  response.content.forEach((item, index) => {
    assert(item.type, `Content item ${index} should have a type`);
    assert(["text", "image", "resource"].includes(item.type), 
      `Content item ${index} should have valid type`);
  });
}

/**
 * Run a command and capture output for testing
 */
export function runCommand(command, args = [], options = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      stdio: ["pipe", "pipe", "pipe"],
      ...options
    });

    let stdout = "";
    let stderr = "";

    child.stdout.on("data", (data) => {
      stdout += data.toString();
    });

    child.stderr.on("data", (data) => {
      stderr += data.toString();
    });

    child.on("close", (code) => {
      resolve({
        code,
        stdout,
        stderr
      });
    });

    child.on("error", reject);

    // Set a timeout for long-running commands
    const timeout = options.timeout || 5000;
    setTimeout(() => {
      child.kill();
      reject(new Error(`Command timed out after ${timeout}ms`));
    }, timeout);
  });
}

/**
 * Sleep for testing async operations
 */
export function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Assert that an error has the expected message pattern
 */
export function assertErrorMessage(error, expectedPattern) {
  assert(error instanceof Error, "Should be an Error instance");
  if (typeof expectedPattern === "string") {
    assert(error.message.includes(expectedPattern), 
      `Error message "${error.message}" should contain "${expectedPattern}"`);
  } else if (expectedPattern instanceof RegExp) {
    assert(expectedPattern.test(error.message), 
      `Error message "${error.message}" should match pattern ${expectedPattern}`);
  }
}

/**
 * Test data factory for VEX statements
 */
export const testData = {
  validVexRequest: {
    product: "example/product@1.0.0",
    vulnerability: "CVE-2024-12345",
    status: "not_affected",
    justification: "component_not_present",
    impact_statement: "This component is not present in our product",
    author: "security-team@example.com"
  },
  
  validAffectedRequest: {
    product: "example/product@1.0.0",
    vulnerability: "CVE-2024-12345",
    status: "affected",
    action_statement: "Update to version 2.0.0 or later",
    author: "security-team@example.com"
  },
  
  invalidVexRequests: [
    { // Missing product
      vulnerability: "CVE-2024-12345",
      status: "not_affected"
    },
    { // Invalid status
      product: "example/product@1.0.0",
      vulnerability: "CVE-2024-12345",
      status: "invalid_status"
    },
    { // Invalid CVE format
      product: "example/product@1.0.0",
      vulnerability: "INVALID-CVE",
      status: "not_affected"
    }
  ]
};
