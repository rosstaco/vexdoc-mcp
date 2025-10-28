package tools

import (
	"context"
	"fmt"

	"github.com/rosstaco/vexdoc-mcp/internal/vex"
	"github.com/rosstaco/vexdoc-mcp/pkg/api"
)

// VEXMergeTool implements the merge_vex_documents MCP tool
type VEXMergeTool struct {
	client *vex.Client
}

// NewVEXMergeTool creates a new VEX merge tool
func NewVEXMergeTool(client *vex.Client) *VEXMergeTool {
	return &VEXMergeTool{client: client}
}

// Name returns the tool name
func (t *VEXMergeTool) Name() string {
	return "merge_vex_documents"
}

// Description returns the tool description
func (t *VEXMergeTool) Description() string {
	return "Merge and consolidate multiple VEX documents into a unified security assessment report. This tool can merge vulnerability statements from different sources, teams, or vendors into a single authoritative VEX document. Supports filtering by products or vulnerabilities."
}

// InputSchema returns the JSON schema for tool input
func (t *VEXMergeTool) InputSchema() *api.JSONSchema {
	return &api.JSONSchema{
		Type: "object",
		Properties: map[string]*api.JSONSchema{
			"documents": {
				Type:        "array",
				Description: "Collection of VEX documents to merge from different sources (vendors, teams, previous assessments). Each must be a complete OpenVEX-formatted document.",
				Items: &api.JSONSchema{
					Type:        "object",
					Description: "Complete OpenVEX document containing vulnerability assessments. Must include @context for format version, statements array with vulnerability assessments, and document metadata.",
				},
			},
			"author": {
				Type:        "string",
				Description: "Security analyst, team, or organization responsible for this vulnerability assessment (e.g., security-team@company.com, John Doe, ACME Security Team)",
			},
			"author_role": {
				Type:        "string",
				Description: "Role or title of the person creating the merged document (e.g., 'Security Engineer', 'Vulnerability Manager', 'CISO')",
			},
			"id": {
				Type:        "string",
				Description: "Custom identifier for the new merged VEX document. If not provided, a unique ID will be automatically generated.",
			},
			"products": {
				Type:        "array",
				Description: "Filter merge to only include vulnerability statements for these specific products. Useful for creating product-specific security reports.",
				Items: &api.JSONSchema{
					Type:        "string",
					Description: "Product identifier in PURL format",
				},
			},
			"vulnerabilities": {
				Type:        "array",
				Description: "Filter merge to only include statements for these specific vulnerabilities. Useful for creating vulnerability-specific impact reports across multiple products.",
				Items: &api.JSONSchema{
					Type:        "string",
					Description: "Security vulnerability identifier from CVE, GHSA, or other vulnerability databases",
				},
			},
		},
		Required: []string{"documents"},
	}
}

// Execute executes the tool with the given arguments
func (t *VEXMergeTool) Execute(ctx context.Context, args map[string]interface{}) (*api.ToolResult, error) {
	// Parse input
	input, err := parseMergeInput(args)
	if err != nil {
		return errorResult(fmt.Sprintf("Error: %s", err.Error())), nil
	}

	// Merge VEX documents (no context needed with simplified client)
	doc, err := t.client.MergeDocuments(input)
	if err != nil {
		return errorResult(fmt.Sprintf("Error: %s", err.Error())), nil
	}

	// Format output as JSON
	output, err := formatVEXDocument(doc)
	if err != nil {
		return errorResult(fmt.Sprintf("Error: failed to format VEX document: %s", err.Error())), nil
	}

	return &api.ToolResult{
		Content: []api.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("VEX documents merged successfully:\n\n%s", output),
			},
		},
	}, nil
}

// parseMergeInput parses and validates merge tool arguments
func parseMergeInput(args map[string]interface{}) (*vex.MergeInput, error) {
	input := &vex.MergeInput{}

	// Required: documents array
	docsInterface, ok := args["documents"]
	if !ok {
		return nil, fmt.Errorf("documents field is required")
	}

	docsArray, ok := docsInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("documents must be an array")
	}

	// Convert each document to map[string]interface{}
	for i, docInterface := range docsArray {
		docMap, ok := docInterface.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("document %d must be a valid JSON object", i+1)
		}
		input.Documents = append(input.Documents, docMap)
	}

	// Optional fields
	if author, ok := args["author"].(string); ok {
		input.Author = author
	}

	if authorRole, ok := args["author_role"].(string); ok {
		input.AuthorRole = authorRole
	}

	if id, ok := args["id"].(string); ok {
		input.ID = id
	}

	// Optional products filter
	if productsInterface, ok := args["products"]; ok {
		if productsArray, ok := productsInterface.([]interface{}); ok {
			for _, p := range productsArray {
				if product, ok := p.(string); ok {
					input.Products = append(input.Products, product)
				}
			}
		}
	}

	// Optional vulnerabilities filter
	if vulnsInterface, ok := args["vulnerabilities"]; ok {
		if vulnsArray, ok := vulnsInterface.([]interface{}); ok {
			for _, v := range vulnsArray {
				if vuln, ok := v.(string); ok {
					input.Vulnerabilities = append(input.Vulnerabilities, vuln)
				}
			}
		}
	}

	return input, nil
}
