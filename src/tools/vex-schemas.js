// Shared VEX schema definitions for reuse across tools

// ============================================================================
// CONSTANTS - Security allowed values to prevent injection
// ============================================================================

export const ALLOWED_STATUSES = ["not_affected", "affected", "fixed", "under_investigation"];
export const ALLOWED_JUSTIFICATIONS = [
  "component_not_present",
  "vulnerable_code_not_present", 
  "vulnerable_code_not_in_execute_path",
  "vulnerable_code_cannot_be_controlled_by_adversary",
  "inline_mitigations_already_exist"
];

// Document processing limits for security and performance
export const MAX_MERGE_DOCUMENTS = 20;

// ============================================================================
// VALIDATION FUNCTIONS - Common VEX validation utilities
// ============================================================================

export function validateVulnerabilityFormat(vulnerability) {
  const pattern = /^(CVE-\d{4}-\d+|GHSA-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}|[A-Z]+-\d+-\d+)$/i;
  return pattern.test(vulnerability.trim());
}

export function validateStatus(status) {
  return ALLOWED_STATUSES.includes(status);
}

export function validateJustification(justification) {
  return !justification || ALLOWED_JUSTIFICATIONS.includes(justification);
}

// ============================================================================
// PROPERTY DEFINITIONS - Reusable schema property objects
// ============================================================================

// Base properties for VEX statements
export const VEX_STATEMENT_PROPERTIES = {
  product: {
    type: "string",
    description: "Software product identifier using PURL (Package URL) format, e.g., pkg:npm/lodash@4.17.21, pkg:docker/nginx@1.20.1, pkg:apk/wolfi/git@2.39.0-r1?arch=x86_64",
    maxLength: 1000
  },
  vulnerability: {
    type: "string",
    description: "Security vulnerability identifier from CVE, GHSA, or other vulnerability databases (e.g., CVE-2023-1234, GHSA-xxxx-xxxx-xxxx)",
    pattern: "^(CVE-\\d{4}-\\d+|GHSA-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}|[A-Z]+-\\d+-\\d+)$",
    maxLength: 50
  },
  status: {
    type: "string",
    enum: ALLOWED_STATUSES,
    description: "Assessment of how the vulnerability affects this product: not_affected (product is safe), affected (vulnerable), fixed (patched), under_investigation (being analyzed)"
  },
  justification: {
    type: "string",
    enum: ALLOWED_JUSTIFICATIONS,
    description: "Technical reason why a product is not affected by the vulnerability (required when status=not_affected): component_not_present, vulnerable_code_not_present, vulnerable_code_not_in_execute_path, vulnerable_code_cannot_be_controlled_by_adversary, inline_mitigations_already_exist"
  },
  impact_statement: {
    type: "string",
    description: "Detailed technical explanation of why the vulnerability cannot be exploited in this product context (used with status=not_affected)",
    maxLength: 1000
  },
  action_statement: {
    type: "string",
    description: "Recommended remediation actions for affected products, such as version upgrades, configuration changes, or workarounds (used with status=affected)",
    maxLength: 1000
  },
  author: {
    type: "string",
    description: "Security analyst, team, or organization responsible for this vulnerability assessment (e.g., security-team@company.com, John Doe, ACME Security Team)",
    maxLength: 200
  }
};

// ============================================================================
// DOCUMENT SCHEMAS - Schema definitions for complete document structures
// ============================================================================

// Schema for complete VEX documents (used in merge operations)
export const VEX_DOCUMENT_SCHEMA = {
  type: "object",
  description: "Complete OpenVEX document containing vulnerability assessments. Must include @context for format version, statements array with vulnerability assessments, and document metadata.",
  properties: {
    "@context": {
      type: "string",
      description: "OpenVEX specification version URL (e.g., https://openvex.dev/ns/v0.2.0)"
    },
    "@id": {
      type: "string",
      description: "Globally unique identifier for this VEX document (URI format recommended)"
    },
    author: {
      type: "string",
      description: "Person, team, or organization who created this VEX document"
    },
    timestamp: {
      type: "string",
      description: "When this VEX document was created or last updated (ISO 8601 format)"
    },
    version: {
      type: "number",
      description: "Version number of this VEX document for tracking updates"
    },
    statements: {
      type: "array",
      description: "Collection of vulnerability assessment statements, each linking specific products to specific vulnerabilities with impact status",
      items: {
        type: "object",
        description: "Individual vulnerability assessment statement"
      }
    }
  }
};

// Additional properties specific to merge operations
export const VEX_MERGE_PROPERTIES = {
  documents: {
    type: "array",
    items: VEX_DOCUMENT_SCHEMA,
    description: "Collection of VEX documents to merge from different sources (vendors, teams, previous assessments). Each must be a complete OpenVEX-formatted document.",
    minItems: 2,
    maxItems: MAX_MERGE_DOCUMENTS // Reasonable limit to prevent resource abuse
  },
  author: VEX_STATEMENT_PROPERTIES.author,
  author_role: {
    type: "string",
    description: "Role or title of the person creating the merged document (e.g., 'Security Engineer', 'Vulnerability Manager', 'CISO')",
    maxLength: 200
  },
  id: {
    type: "string",
    description: "Custom identifier for the new merged VEX document. If not provided, a unique ID will be automatically generated.",
    maxLength: 500
  },
  products: {
    type: "array",
    items: {
      type: "string",
      description: "Product identifier in PURL format"
    },
    description: "Filter merge to only include vulnerability statements for these specific products. Useful for creating product-specific security reports."
  },
  vulnerabilities: {
    type: "array", 
    items: VEX_STATEMENT_PROPERTIES.vulnerability,
    description: "Filter merge to only include statements for these specific vulnerabilities. Useful for creating vulnerability-specific impact reports across multiple products."
  }
};

// ============================================================================
// TOOL SCHEMAS - Complete input schemas for MCP tools
// ============================================================================

// Complete schema for VEX statement creation tool
export const VEX_CREATE_SCHEMA = {
  type: "object",
  properties: VEX_STATEMENT_PROPERTIES,
  required: ["product", "vulnerability", "status"]
};

// Complete schema for VEX document merge tool
export const VEX_MERGE_SCHEMA = {
  type: "object",
  properties: VEX_MERGE_PROPERTIES,
  required: ["documents"]
};
