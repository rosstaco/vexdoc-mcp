import { createVexStatementTool, handleCreateVexStatement } from "./vex-create.js";
import { mergeVexDocumentsTool, handleMergeVexDocuments } from "./vex-merge.js";

// Export all tools
export const tools = [
  createVexStatementTool,
  mergeVexDocumentsTool
];

// Export tool handlers
export const toolHandlers = {
  create_vex_statement: handleCreateVexStatement,
  merge_vex_documents: handleMergeVexDocuments
};
