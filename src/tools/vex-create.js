import { spawn } from "child_process";
import { 
  ALLOWED_STATUSES, 
  ALLOWED_JUSTIFICATIONS, 
  VEX_CREATE_SCHEMA,
  validateVulnerabilityFormat,
  validateStatus,
  validateJustification
} from "./vex-schemas.js";

// Security: Validate and sanitize input parameters
function validateAndSanitizeInput(args) {
  const { product, vulnerability, status, justification, impact_statement, action_statement, author } = args;
  
  // Validate required parameters
  if (!product || typeof product !== "string" || product.trim().length === 0) {
    throw new Error("Product parameter is required and must be a non-empty string");
  }
  
  if (!vulnerability || typeof vulnerability !== "string" || vulnerability.trim().length === 0) {
    throw new Error("Vulnerability parameter is required and must be a non-empty string");
  }
  
  if (!status || !validateStatus(status)) {
    throw new Error(`Status must be one of: ${ALLOWED_STATUSES.join(", ")}`);
  }
  
  // Validate vulnerability format (CVE or similar)
  if (!validateVulnerabilityFormat(vulnerability)) {
    throw new Error("Vulnerability must be in valid format (e.g., CVE-2023-1234, GHSA-xxxx-xxxx-xxxx)");
  }
  
  // Validate justification if provided
  if (!validateJustification(justification)) {
    throw new Error(`Justification must be one of: ${ALLOWED_JUSTIFICATIONS.join(", ")}`);
  }
  
  // Security: Prevent command injection by checking for dangerous characters
  const dangerousChars = /[;&|`$(){}[\]<>'"\\]/;
  const stringParams = { product, vulnerability, impact_statement, action_statement, author };
  
  for (const [key, value] of Object.entries(stringParams)) {
    if (value && dangerousChars.test(value)) {
      throw new Error(`${key} parameter contains potentially dangerous characters`);
    }
  }
  
  // Security: Limit string lengths to prevent buffer overflows
  const maxLength = 1000;
  const stringLengthChecks = [
    { name: "product", value: product, max: maxLength },
    { name: "vulnerability", value: vulnerability, max: 50 },
    { name: "impact_statement", value: impact_statement, max: maxLength },
    { name: "action_statement", value: action_statement, max: maxLength },
    { name: "author", value: author, max: 200 }
  ];
  
  for (const { name, value, max } of stringLengthChecks) {
    if (value && value.length > max) {
      throw new Error(`${name} parameter exceeds maximum length of ${max} characters`);
    }
  }
  
  return {
    product: product.trim(),
    vulnerability: vulnerability.trim(),
    status,
    justification,
    impact_statement: impact_statement?.trim(),
    action_statement: action_statement?.trim(), 
    author: author?.trim()
  };
}

export const createVexStatementTool = {
  name: "create_vex_statement",
  description: "Generate VEX (Vulnerability Exploitability eXchange) statements to document security vulnerability assessments for software products. Creates OpenVEX-compliant JSON documents that specify whether products are affected by specific vulnerabilities.",
  inputSchema: VEX_CREATE_SCHEMA
};

export async function handleCreateVexStatement(args) {
  try {
    // Security: Validate and sanitize all inputs first
    const sanitizedArgs = validateAndSanitizeInput(args);
    
    const { 
      product, 
      vulnerability, 
      status, 
      justification, 
      impact_statement, 
      action_statement, 
      author
    } = sanitizedArgs;

    // Validate required justification for not_affected status
    if (status === "not_affected" && !justification) {
      return {
        content: [
          {
            type: "text",
            text: "Error: Justification is required when status is 'not_affected'"
          }
        ],
        isError: true
      };
    }

    // Rest of the function continues here...
    return await executeVexCtl(product, vulnerability, status, justification, impact_statement, action_statement, author);
    
  } catch (error) {
    return {
      content: [
        {
          type: "text",
          text: `Error: ${error.message}`
        }
      ],
      isError: true
    };
  }
}

async function executeVexCtl(product, vulnerability, status, justification, impact_statement, action_statement, author) {
  const vexctlArgs = ["create"];
  
  // Security: Each argument is passed separately to prevent injection
  vexctlArgs.push("--product", product);
  vexctlArgs.push("--vuln", vulnerability);
  vexctlArgs.push("--status", status);
  
  // Add status-specific arguments
  if (status === "not_affected") {
    if (justification) {
      vexctlArgs.push("--justification", justification);
    }
    if (impact_statement) {
      vexctlArgs.push("--impact-statement", impact_statement);
    }
  } else if (status === "affected") {
    if (action_statement) {
      vexctlArgs.push("--action-statement", action_statement);
    }
  }
  
  // Author can be added for any status
  if (author) {
    vexctlArgs.push("--author", author);
  }

  return new Promise((resolve, reject) => {
    // Security: Use spawn with argument array to prevent shell injection
    const vexctl = spawn("vexctl", vexctlArgs, {
      stdio: ["ignore", "pipe", "pipe"], // Don't pass stdin
      shell: false, // Explicitly disable shell to prevent injection
      timeout: 30000 // 30 second timeout
    });
    
    let stdout = "";
    let stderr = "";
    
    vexctl.stdout.on("data", (data) => {
      stdout += data.toString();
      // Security: Limit output size to prevent DoS
      if (stdout.length > 100000) { // 100KB limit
        vexctl.kill("SIGTERM");
        reject(new Error("Command output exceeded size limit"));
        return;
      }
    });
    
    vexctl.stderr.on("data", (data) => {
      stderr += data.toString();
      // Security: Limit error output size
      if (stderr.length > 10000) { // 10KB limit
        vexctl.kill("SIGTERM");
        reject(new Error("Command error output exceeded size limit"));
        return;
      }
    });
    
    vexctl.on("close", (code) => {
      if (code !== 0) {
        // Security: Sanitize error message to prevent info disclosure
        const sanitizedError = stderr.replace(/\/[^\s]*vexctl[^\s]*/g, "vexctl");
        reject(new Error(`vexctl failed with exit code ${code}: ${sanitizedError}`));
        return;
      }
      
      // Parse and return the VEX document as JSON
      try {
        const vexDocument = JSON.parse(stdout);
        
        resolve({
          content: [
            {
              type: "text",
              text: `VEX statement created successfully:\n\n${JSON.stringify(vexDocument, null, 2)}`
            }
          ]
        });
      } catch {
        // If JSON parsing fails, return raw output
        resolve({
          content: [
            {
              type: "text",
              text: `VEX statement created successfully:\n\n${stdout}`
            }
          ]
        });
      }
    });
    
    vexctl.on("error", (error) => {
      reject(new Error(`Failed to execute vexctl: ${error.message}`));
    });
    
    // Security: Set a timeout for the process
    setTimeout(() => {
      if (!vexctl.killed) {
        vexctl.kill("SIGTERM");
        reject(new Error("Command execution timed out"));
      }
    }, 30000);
  });
}
