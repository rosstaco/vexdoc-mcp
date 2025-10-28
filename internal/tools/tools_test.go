package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/rosstaco/vexdoc-mcp/internal/vex"
)

func TestVEXCreateTool_Name(t *testing.T) {
	client := vex.NewClient("test-author")
	tool := NewVEXCreateTool(client)

	if tool.Name() != "create_vex_statement" {
		t.Errorf("Name() = %v, want create_vex_statement", tool.Name())
	}
}

func TestVEXCreateTool_Description(t *testing.T) {
	client := vex.NewClient("test-author")
	tool := NewVEXCreateTool(client)

	desc := tool.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	if !strings.Contains(desc, "VEX") {
		t.Error("Description() should mention VEX")
	}
}

func TestVEXCreateTool_InputSchema(t *testing.T) {
	client := vex.NewClient("test-author")
	tool := NewVEXCreateTool(client)

	schema := tool.InputSchema()
	if schema == nil {
		t.Fatal("InputSchema() returned nil")
	}

	if schema.Type != "object" {
		t.Errorf("Schema type = %v, want object", schema.Type)
	}

	// Check required fields
	expectedRequired := []string{"product", "vulnerability", "status"}
	if len(schema.Required) != len(expectedRequired) {
		t.Errorf("Required fields count = %v, want %v", len(schema.Required), len(expectedRequired))
	}
	for _, req := range expectedRequired {
		found := false
		for _, field := range schema.Required {
			if field == req {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required field %v not found in schema", req)
		}
	}

	// Check properties exist
	expectedProps := []string{"product", "vulnerability", "status", "justification", "impact_statement", "action_statement", "author"}
	for _, prop := range expectedProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Property %v not found in schema", prop)
		}
	}
}

func TestVEXCreateTool_Execute_Success(t *testing.T) {
	client := vex.NewClient("test-author")
	tool := NewVEXCreateTool(client)
	ctx := context.Background()

	tests := []struct {
		name string
		args map[string]interface{}
	}{
		{
			name: "minimal required fields",
			args: map[string]interface{}{
				"product":       "pkg:npm/lodash@4.17.21",
				"vulnerability": "CVE-2023-1234",
				"status":        "not_affected",
				"justification": "component_not_present",
			},
		},
		{
			name: "with action statement",
			args: map[string]interface{}{
				"product":          "pkg:npm/express@4.18.0",
				"vulnerability":    "CVE-2023-5678",
				"status":           "affected",
				"action_statement": "Update to version 5.0.0",
				"author":           "security-team",
			},
		},
		{
			name: "fixed status",
			args: map[string]interface{}{
				"product":       "pkg:npm/react@17.0.0",
				"vulnerability": "CVE-2023-9999",
				"status":        "fixed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(ctx, tt.args)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if result == nil {
				t.Fatal("Execute() returned nil result")
			}
			if result.IsError {
				t.Errorf("Execute() returned error result: %v", result.Content[0].Text)
			}
			if len(result.Content) == 0 {
				t.Error("Execute() returned empty content")
			}
			if result.Content[0].Type != "text" {
				t.Errorf("Content type = %v, want text", result.Content[0].Type)
			}
			if !strings.Contains(result.Content[0].Text, "successfully") {
				t.Error("Success message should contain 'successfully'")
			}
			if !strings.Contains(result.Content[0].Text, "@context") {
				t.Error("Result should contain VEX document JSON")
			}
		})
	}
}

func TestVEXCreateTool_Execute_ValidationErrors(t *testing.T) {
	client := vex.NewClient("test-author")
	tool := NewVEXCreateTool(client)
	ctx := context.Background()

	tests := []struct {
		name            string
		args            map[string]interface{}
		wantErrContains string
	}{
		{
			name: "missing product",
			args: map[string]interface{}{
				"vulnerability": "CVE-2023-1234",
				"status":        "not_affected",
				"justification": "component_not_present",
			},
			wantErrContains: "product",
		},
		{
			name: "missing vulnerability",
			args: map[string]interface{}{
				"product":       "pkg:npm/lodash@4.17.21",
				"status":        "not_affected",
				"justification": "component_not_present",
			},
			wantErrContains: "vulnerability",
		},
		{
			name: "missing status",
			args: map[string]interface{}{
				"product":       "pkg:npm/lodash@4.17.21",
				"vulnerability": "CVE-2023-1234",
				"justification": "component_not_present",
			},
			wantErrContains: "status",
		},
		{
			name: "invalid product type",
			args: map[string]interface{}{
				"product":       123, // Should be string
				"vulnerability": "CVE-2023-1234",
				"status":        "not_affected",
			},
			wantErrContains: "product",
		},
		{
			name: "product too long",
			args: map[string]interface{}{
				"product":       strings.Repeat("a", 1001),
				"vulnerability": "CVE-2023-1234",
				"status":        "not_affected",
				"justification": "component_not_present",
			},
			wantErrContains: "maximum length",
		},
		{
			name: "invalid status value",
			args: map[string]interface{}{
				"product":       "pkg:npm/lodash@4.17.21",
				"vulnerability": "CVE-2023-1234",
				"status":        "invalid_status",
			},
			wantErrContains: "status",
		},
		{
			name: "not_affected without justification or impact",
			args: map[string]interface{}{
				"product":       "pkg:npm/lodash@4.17.21",
				"vulnerability": "CVE-2023-1234",
				"status":        "not_affected",
			},
			wantErrContains: "justification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(ctx, tt.args)

			// Tool should return error result, not error
			if err != nil {
				t.Fatalf("Execute() unexpected error = %v", err)
			}
			if result == nil {
				t.Fatal("Execute() returned nil result")
			}
			if !result.IsError {
				t.Error("Execute() should return error result for invalid input")
			}
			if len(result.Content) == 0 {
				t.Fatal("Execute() returned empty content")
			}

			errorText := result.Content[0].Text
			if !strings.Contains(errorText, tt.wantErrContains) {
				t.Errorf("Error text = %v, want to contain %v", errorText, tt.wantErrContains)
			}
		})
	}
}

func TestVEXMergeTool_Name(t *testing.T) {
	client := vex.NewClient("test-author")
	tool := NewVEXMergeTool(client)

	if tool.Name() != "merge_vex_documents" {
		t.Errorf("Name() = %v, want merge_vex_documents", tool.Name())
	}
}

func TestVEXMergeTool_Description(t *testing.T) {
	client := vex.NewClient("test-author")
	tool := NewVEXMergeTool(client)

	desc := tool.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
	if !strings.Contains(desc, "merge") {
		t.Error("Description() should mention merge")
	}
}

func TestVEXMergeTool_InputSchema(t *testing.T) {
	client := vex.NewClient("test-author")
	tool := NewVEXMergeTool(client)

	schema := tool.InputSchema()
	if schema == nil {
		t.Fatal("InputSchema() returned nil")
	}

	if schema.Type != "object" {
		t.Errorf("Schema type = %v, want object", schema.Type)
	}

	// Check required fields
	if len(schema.Required) != 1 || schema.Required[0] != "documents" {
		t.Errorf("Required fields = %v, want [documents]", schema.Required)
	}

	// Check properties exist
	expectedProps := []string{"documents", "author", "author_role", "id", "products", "vulnerabilities"}
	for _, prop := range expectedProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Property %v not found in schema", prop)
		}
	}
}

func TestVEXMergeTool_Execute_Success(t *testing.T) {
	client := vex.NewClient("test-author")
	tool := NewVEXMergeTool(client)
	ctx := context.Background()

	doc1 := map[string]interface{}{
		"@context":  "https://openvex.dev/ns",
		"@id":       "doc1",
		"author":    "author1",
		"version":   1,
		"timestamp": "2023-01-01T00:00:00Z",
		"statements": []interface{}{
			map[string]interface{}{
				"vulnerability": map[string]interface{}{"name": "CVE-2023-1234"},
				"products":      []interface{}{map[string]interface{}{"@id": "pkg:npm/lodash@4.17.21"}},
				"status":        "not_affected",
				"justification": "component_not_present",
			},
		},
	}

	doc2 := map[string]interface{}{
		"@context":  "https://openvex.dev/ns",
		"@id":       "doc2",
		"author":    "author2",
		"version":   1,
		"timestamp": "2023-01-02T00:00:00Z",
		"statements": []interface{}{
			map[string]interface{}{
				"vulnerability": map[string]interface{}{"name": "CVE-2023-5678"},
				"products":      []interface{}{map[string]interface{}{"@id": "pkg:npm/express@4.18.0"}},
				"status":        "affected",
			},
		},
	}

	tests := []struct {
		name string
		args map[string]interface{}
	}{
		{
			name: "basic merge",
			args: map[string]interface{}{
				"documents": []interface{}{doc1, doc2},
			},
		},
		{
			name: "merge with metadata",
			args: map[string]interface{}{
				"documents":   []interface{}{doc1, doc2},
				"author":      "merger",
				"author_role": "Security Lead",
				"id":          "merged-doc",
			},
		},
		{
			name: "merge with product filter",
			args: map[string]interface{}{
				"documents": []interface{}{doc1, doc2},
				"products":  []interface{}{"pkg:npm/lodash@4.17.21"},
			},
		},
		{
			name: "merge with vulnerability filter",
			args: map[string]interface{}{
				"documents":       []interface{}{doc1, doc2},
				"vulnerabilities": []interface{}{"CVE-2023-1234"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(ctx, tt.args)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if result == nil {
				t.Fatal("Execute() returned nil result")
			}
			if result.IsError {
				t.Errorf("Execute() returned error result: %v", result.Content[0].Text)
			}
			if len(result.Content) == 0 {
				t.Error("Execute() returned empty content")
			}
			if !strings.Contains(result.Content[0].Text, "successfully") {
				t.Error("Success message should contain 'successfully'")
			}
			if !strings.Contains(result.Content[0].Text, "@context") {
				t.Error("Result should contain merged VEX document JSON")
			}
		})
	}
}

func TestVEXMergeTool_Execute_ValidationErrors(t *testing.T) {
	client := vex.NewClient("test-author")
	tool := NewVEXMergeTool(client)
	ctx := context.Background()

	doc := map[string]interface{}{
		"@context":   "https://openvex.dev/ns",
		"@id":        "doc1",
		"author":     "author1",
		"version":    1,
		"timestamp":  "2023-01-01T00:00:00Z",
		"statements": []interface{}{},
	}

	tests := []struct {
		name            string
		args            map[string]interface{}
		wantErrContains string
	}{
		{
			name:            "missing documents",
			args:            map[string]interface{}{},
			wantErrContains: "documents",
		},
		{
			name: "too few documents",
			args: map[string]interface{}{
				"documents": []interface{}{doc},
			},
			wantErrContains: "at least",
		},
		{
			name: "too many documents",
			args: map[string]interface{}{
				"documents": func() []interface{} {
					docs := make([]interface{}, 21)
					for i := range docs {
						docs[i] = map[string]interface{}{"@context": "test", "statements": []interface{}{}}
					}
					return docs
				}(),
			},
			wantErrContains: "maximum",
		},
		{
			name: "invalid document type",
			args: map[string]interface{}{
				"documents": []interface{}{"not a map", "also not a map"},
			},
			wantErrContains: "JSON object",
		},
		{
			name: "document missing context",
			args: map[string]interface{}{
				"documents": []interface{}{
					map[string]interface{}{"statements": []interface{}{}},
					map[string]interface{}{"@context": "https://openvex.dev/ns", "statements": []interface{}{}},
				},
			},
			wantErrContains: "@context",
		},
		{
			name: "document missing statements",
			args: map[string]interface{}{
				"documents": []interface{}{
					map[string]interface{}{"@context": "https://openvex.dev/ns"},
					map[string]interface{}{"@context": "https://openvex.dev/ns", "statements": []interface{}{}},
				},
			},
			wantErrContains: "statements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(ctx, tt.args)

			// Tool should return error result, not error
			if err != nil {
				t.Fatalf("Execute() unexpected error = %v", err)
			}
			if result == nil {
				t.Fatal("Execute() returned nil result")
			}
			if !result.IsError {
				t.Error("Execute() should return error result for invalid input")
			}
			if len(result.Content) == 0 {
				t.Fatal("Execute() returned empty content")
			}

			errorText := result.Content[0].Text
			if !strings.Contains(errorText, tt.wantErrContains) {
				t.Errorf("Error text = %v, want to contain %v", errorText, tt.wantErrContains)
			}
		})
	}
}

func TestFormatVEXDocument(t *testing.T) {
	doc := map[string]interface{}{
		"@context": "https://openvex.dev/ns",
		"@id":      "test-doc",
		"version":  1,
	}

	output, err := formatVEXDocument(doc)
	if err != nil {
		t.Fatalf("formatVEXDocument() error = %v", err)
	}

	if output == "" {
		t.Error("formatVEXDocument() returned empty string")
	}

	// Should be valid JSON
	if !strings.Contains(output, "@context") {
		t.Error("Output should contain @context")
	}
	if !strings.Contains(output, "test-doc") {
		t.Error("Output should contain document ID")
	}
}

func TestErrorResult(t *testing.T) {
	message := "test error message"
	result := errorResult(message)

	if result == nil {
		t.Fatal("errorResult() returned nil")
	}
	if !result.IsError {
		t.Error("errorResult() should set IsError to true")
	}
	if len(result.Content) != 1 {
		t.Fatalf("errorResult() content length = %v, want 1", len(result.Content))
	}
	if result.Content[0].Type != "text" {
		t.Errorf("errorResult() content type = %v, want text", result.Content[0].Type)
	}
	if result.Content[0].Text != message {
		t.Errorf("errorResult() content text = %v, want %v", result.Content[0].Text, message)
	}
}
