import { createVexStatementTool, handleCreateVexStatement } from "./vexctl.js";

// Export all tools
export const tools = [
  createVexStatementTool
];

// Export tool handlers
export const toolHandlers = {
  create_vex_statement: handleCreateVexStatement
};
