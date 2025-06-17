import { spawn } from "child_process";
import { writeFile, unlink, mkdtemp, readdir, rmdir } from "fs/promises";
import { join } from "path";
import { tmpdir } from "os";
import { 
  VEX_MERGE_SCHEMA,
  MAX_MERGE_DOCUMENTS
} from "./vex-schemas.js";

export const mergeVexDocumentsTool = {
  name: "merge_vex_documents",
  description: "Merge and consolidate multiple VEX documents into a unified security assessment report. This tool can merge vulnerability statements from different sources, teams, or vendors into a single authoritative VEX document. Supports filtering by products or vulnerabilities.",
  inputSchema: VEX_MERGE_SCHEMA
};

export async function handleMergeVexDocuments(args) {
  const tempFiles = [];
  let tempDir = null;
  
  try {
    // Security: Validate and sanitize all inputs first
    const sanitizedArgs = validateAndSanitizeMergeInput(args);
    
    const { 
      documents,
      author,
      author_role,
      id,
      products,
      vulnerabilities
    } = sanitizedArgs;

    // Create secure temporary directory
    tempDir = await mkdtemp(join(tmpdir(), "vexctl-merge-"));
    
    // Write each document to a temporary file
    for (let i = 0; i < documents.length; i++) {
      const tempFile = join(tempDir, `vex-doc-${i}.json`);
      await writeFile(tempFile, JSON.stringify(documents[i], null, 2));
      tempFiles.push(tempFile);
    }
    
    // Execute vexctl merge
    return await executeVexCtlMerge(tempDir, {
      author,
      author_role,
      id,
      products,
      vulnerabilities
    });
    
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
  } finally {
    // Always clean up temporary files
    await cleanupTempFiles(tempFiles, tempDir);
  }
}

function validateAndSanitizeMergeInput(args) {
  const { documents, author, author_role, id, products, vulnerabilities } = args;
  
  // Validate required parameters
  if (!documents || !Array.isArray(documents) || documents.length < 2) {
    throw new Error("At least 2 VEX documents are required for merging");
  }
  
  if (documents.length > MAX_MERGE_DOCUMENTS) { // Reasonable limit for merge operations
    throw new Error(`Maximum of ${MAX_MERGE_DOCUMENTS} documents can be merged at once`);
  }
  
  // Validate each document is a valid object
  documents.forEach((doc, index) => {
    if (!doc || typeof doc !== "object") {
      throw new Error(`Document ${index + 1} must be a valid JSON object`);
    }
    
    // Basic VEX document structure validation
    if (!doc["@context"] || !doc.statements) {
      throw new Error(`Document ${index + 1} must be a valid VEX document with @context and statements`);
    }
  });
  
  // Security: Prevent command injection by checking for dangerous characters
  const dangerousChars = /[;&|`$(){}[\]<>'"\\]/;
  const stringParams = { author, author_role, id };
  
  for (const [key, value] of Object.entries(stringParams)) {
    if (value && dangerousChars.test(value)) {
      throw new Error(`${key} parameter contains potentially dangerous characters`);
    }
  }
  
  return {
    documents,
    author: author?.trim(),
    author_role: author_role?.trim(),
    id: id?.trim(),
    products: products?.map(p => p.trim()).filter(Boolean),
    vulnerabilities: vulnerabilities?.map(v => v.trim()).filter(Boolean)
  };
}

async function executeVexCtlMerge(tempDir, options) {
  const vexctlArgs = ["merge"];
  
  // Add optional parameters
  if (options.author) {
    vexctlArgs.push("--author", options.author);
  }
  
  if (options.author_role) {
    vexctlArgs.push("--author-role", options.author_role);
  }
  
  if (options.id) {
    vexctlArgs.push("--id", options.id);
  }
  
  if (options.products && options.products.length > 0) {
    options.products.forEach(product => {
      vexctlArgs.push("--product", product);
    });
  }
  
  if (options.vulnerabilities && options.vulnerabilities.length > 0) {
    options.vulnerabilities.forEach(vuln => {
      vexctlArgs.push("--vuln", vuln);
    });
  }
  
  // Add files from temp directory (manually expand the pattern)
  const tempFiles = await readdir(tempDir);
  const vexFiles = tempFiles
    .filter(file => file.startsWith("vex-doc-") && file.endsWith(".json"))
    .map(file => join(tempDir, file));
  
  vexctlArgs.push(...vexFiles);

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
        reject(new Error(`vexctl merge failed with exit code ${code}: ${sanitizedError}`));
        return;
      }
      
      // Parse and return the merged VEX document as JSON
      try {
        const mergedDocument = JSON.parse(stdout);
        
        resolve({
          content: [
            {
              type: "text",
              text: `VEX documents merged successfully:\n\n${JSON.stringify(mergedDocument, null, 2)}`
            }
          ]
        });
      } catch {
        // If JSON parsing fails, return raw output
        resolve({
          content: [
            {
              type: "text",
              text: `VEX documents merged successfully:\n\n${stdout}`
            }
          ]
        });
      }
    });
    
    vexctl.on("error", (error) => {
      reject(new Error(`Failed to execute vexctl merge: ${error.message}`));
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

async function cleanupTempFiles(tempFiles, tempDir) {
  // Clean up temporary files
  for (const file of tempFiles) {
    try {
      await unlink(file);
    } catch (error) {
      // Log but don't throw - cleanup is best effort
      console.error(`Failed to cleanup temp file ${file}:`, error.message);
    }
  }
  
  // Clean up temporary directory if it exists and is empty
  if (tempDir) {
    try {
      await rmdir(tempDir);
    } catch {
      // Directory might not be empty or might not exist - this is ok
      // Don't log this as it's expected behavior
    }
  }
}
