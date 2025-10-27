package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rosstaco/vexdoc-mcp-go/internal/vex"
	"github.com/rosstaco/vexdoc-mcp-go/pkg/api"
)

// VEXCreateTool implements the create_vex_statement MCP tool
type VEXCreateTool struct {
	client *vex.Client
}

// NewVEXCreateTool creates a new VEX create tool
func NewVEXCreateTool(client *vex.Client) *VEXCreateTool {
	return &VEXCreateTool{client: client}
}

// Name returns the tool name
func (t *VEXCreateTool) Name() string {
	return "create_vex_statement"
}

// Description returns the tool description
func (t *VEXCreateTool) Description() string {
	return "Generate VEX (Vulnerability Exploitability eXchange) statements to document security vulnerability assessments for software products. Creates OpenVEX-compliant JSON documents that specify whether products are affected by specific vulnerabilities."
}

// InputSchema returns the JSON schema for tool input
func (t *VEXCreateTool) InputSchema() *api.JSONSchema {
	return &api.JSONSchema{
		Type: "object",
		Properties: map[string]*api.JSONSchema{
			"product": {
				Type:        "string",
				Description: "Software product identifier using PURL (Package URL) format, e.g., pkg:npm/lodash@4.17.21, pkg:docker/nginx@1.20.1, pkg:apk/wolfi/git@2.39.0-r1?arch=x86_64",
			},
			"vulnerability": {
				Type:        "string",
				Description: "Security vulnerability identifier from CVE, GHSA, or other vulnerability databases (e.g., CVE-2023-1234, GHSA-xxxx-xxxx-xxxx)",
			},
			"status": {
				Type:        "string",
				Description: "Assessment of how the vulnerability affects this product: not_affected (product is safe), affected (vulnerable), fixed (patched), under_investigation (being analyzed)",
				Enum:        []string{"not_affected", "affected", "fixed", "under_investigation"},
			},
			"justification": {
				Type:        "string",
				Description: "Technical reason why a product is not affected by the vulnerability (required when status=not_affected): component_not_present, vulnerable_code_not_present, vulnerable_code_not_in_execute_path, vulnerable_code_cannot_be_controlled_by_adversary, inline_mitigations_already_exist",
				Enum:        []string{"component_not_present", "vulnerable_code_not_present", "vulnerable_code_not_in_execute_path", "vulnerable_code_cannot_be_controlled_by_adversary", "inline_mitigations_already_exist"},
			},
			"impact_statement": {
				Type:        "string",
				Description: "Detailed technical explanation of why the vulnerability cannot be exploited in this product context (used with status=not_affected)",
			},
			"action_statement": {
				Type:        "string",
				Description: "Recommended remediation actions for affected products, such as version upgrades, configuration changes, or workarounds (used with status=affected)",
			},
			"author": {
				Type:        "string",
				Description: "Security analyst, team, or organization responsible for this vulnerability assessment (e.g., security-team@company.com, John Doe, ACME Security Team)",
			},
		},
		Required: []string{"product", "vulnerability", "status"},
	}
}

// Execute runs the tool with the provided arguments
func (t *VEXCreateTool) Execute(ctx context.Context, args map[string]interface{}) (*api.ToolResult, error) {
	// Parse required fields
	product, ok := args["product"].(string)
	if !ok {
		return errorResult("product is required and must be a string"), nil
	}

	vulnerability, ok := args["vulnerability"].(string)
	if !ok {
		return errorResult("vulnerability is required and must be a string"), nil
	}

	status, ok := args["status"].(string)
	if !ok {
		return errorResult("status is required and must be a string"), nil
	}

	// Parse optional fields
	justification, _ := args["justification"].(string)
	impactStatement, _ := args["impact_statement"].(string)
	actionStatement, _ := args["action_statement"].(string)
	author, _ := args["author"].(string)

	// Create VEX statement using simplified client
	doc, err := t.client.CreateStatement(
		product,
		vulnerability,
		status,
		justification,
		impactStatement,
		actionStatement,
		author,
	)
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
				Text: fmt.Sprintf("VEX statement created successfully:\n\n%s", output),
			},
		},
	}, nil
}

// formatVEXDocument formats a VEX document as JSON
func formatVEXDocument(doc interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// errorResult creates an error tool result
func errorResult(message string) *api.ToolResult {
	return &api.ToolResult{
		Content: []api.Content{
			{
				Type: "text",
				Text: message,
			},
		},
		IsError: true,
	}
}
