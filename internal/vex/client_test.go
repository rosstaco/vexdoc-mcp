package vex

import (
	"encoding/json"
	"strings"
	"testing"

	vexlib "github.com/openvex/go-vex/pkg/vex"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name          string
		defaultAuthor string
		wantAuthor    string
	}{
		{
			name:          "with custom author",
			defaultAuthor: "test-author",
			wantAuthor:    "test-author",
		},
		{
			name:          "with empty author uses default",
			defaultAuthor: "",
			wantAuthor:    "vexdoc-mcp-server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.defaultAuthor)
			if client == nil {
				t.Fatal("NewClient() returned nil")
			}
			if client.defaultAuthor != tt.wantAuthor {
				t.Errorf("NewClient() defaultAuthor = %v, want %v", client.defaultAuthor, tt.wantAuthor)
			}
		})
	}
}

func TestCreateStatement_Success(t *testing.T) {
	client := NewClient("test-author")

	tests := []struct {
		name            string
		product         string
		vulnerability   string
		status          string
		justification   string
		impactStatement string
		actionStatement string
		author          string
	}{
		{
			name:          "not_affected with justification",
			product:       "pkg:npm/lodash@4.17.21",
			vulnerability: "CVE-2023-1234",
			status:        "not_affected",
			justification: "component_not_present",
			author:        "security-team",
		},
		{
			name:            "not_affected with impact statement",
			product:         "pkg:npm/express@4.18.0",
			vulnerability:   "CVE-2023-5678",
			status:          "not_affected",
			impactStatement: "The vulnerable code path is not reachable in our deployment",
			author:          "security-team",
		},
		{
			name:            "affected with action statement",
			product:         "pkg:npm/axios@0.21.0",
			vulnerability:   "CVE-2023-9999",
			status:          "affected",
			actionStatement: "Update to version 1.0.0 or later",
			author:          "security-team",
		},
		{
			name:          "fixed status",
			product:       "pkg:npm/react@17.0.0",
			vulnerability: "CVE-2023-1111",
			status:        "fixed",
			author:        "security-team",
		},
		{
			name:          "under_investigation status",
			product:       "pkg:npm/vue@3.0.0",
			vulnerability: "CVE-2023-2222",
			status:        "under_investigation",
			author:        "security-team",
		},
		{
			name:          "with default author",
			product:       "pkg:npm/webpack@5.0.0",
			vulnerability: "CVE-2023-3333",
			status:        "not_affected",
			justification: "vulnerable_code_not_present",
			author:        "", // Should use default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := client.CreateStatement(
				tt.product,
				tt.vulnerability,
				tt.status,
				tt.justification,
				tt.impactStatement,
				tt.actionStatement,
				tt.author,
			)

			if err != nil {
				t.Fatalf("CreateStatement() error = %v", err)
			}
			if doc == nil {
				t.Fatal("CreateStatement() returned nil document")
			}

			// Verify basic structure
			if doc.Context != vexlib.Context {
				t.Errorf("Context = %v, want %v", doc.Context, vexlib.Context)
			}
			if doc.ID == "" {
				t.Error("ID is empty")
			}
			if doc.Version != 1 {
				t.Errorf("Version = %v, want 1", doc.Version)
			}
			if doc.Timestamp == nil {
				t.Error("Timestamp is nil")
			}
			if len(doc.Statements) != 1 {
				t.Errorf("Statements length = %v, want 1", len(doc.Statements))
			}

			// Verify author
			expectedAuthor := tt.author
			if expectedAuthor == "" {
				expectedAuthor = "test-author"
			}
			if doc.Author != expectedAuthor {
				t.Errorf("Author = %v, want %v", doc.Author, expectedAuthor)
			}

			// Verify statement
			stmt := doc.Statements[0]
			if string(stmt.Vulnerability.Name) != tt.vulnerability {
				t.Errorf("Vulnerability = %v, want %v", stmt.Vulnerability.Name, tt.vulnerability)
			}
			if len(stmt.Products) != 1 {
				t.Errorf("Products length = %v, want 1", len(stmt.Products))
			}
			if stmt.Products[0].Component.ID != tt.product {
				t.Errorf("Product = %v, want %v", stmt.Products[0].Component.ID, tt.product)
			}

			// Verify optional fields
			if tt.justification != "" && stmt.Justification == "" {
				t.Errorf("Justification not set")
			}
			if tt.impactStatement != "" && stmt.ImpactStatement != tt.impactStatement {
				t.Errorf("ImpactStatement = %v, want %v", stmt.ImpactStatement, tt.impactStatement)
			}
			if tt.actionStatement != "" && stmt.ActionStatement != tt.actionStatement {
				t.Errorf("ActionStatement = %v, want %v", stmt.ActionStatement, tt.actionStatement)
			}
		})
	}
}

func TestCreateStatement_ValidationErrors(t *testing.T) {
	client := NewClient("test-author")

	tests := []struct {
		name            string
		product         string
		vulnerability   string
		status          string
		justification   string
		impactStatement string
		actionStatement string
		author          string
		wantErrContains string
	}{
		{
			name:            "missing product",
			product:         "",
			vulnerability:   "CVE-2023-1234",
			status:          "not_affected",
			justification:   "component_not_present",
			wantErrContains: "product is required",
		},
		{
			name:            "missing vulnerability",
			product:         "pkg:npm/lodash@4.17.21",
			vulnerability:   "",
			status:          "not_affected",
			justification:   "component_not_present",
			wantErrContains: "vulnerability is required",
		},
		{
			name:            "missing status",
			product:         "pkg:npm/lodash@4.17.21",
			vulnerability:   "CVE-2023-1234",
			status:          "",
			justification:   "component_not_present",
			wantErrContains: "status is required",
		},
		{
			name:            "product too long",
			product:         strings.Repeat("a", 1001),
			vulnerability:   "CVE-2023-1234",
			status:          "not_affected",
			justification:   "component_not_present",
			wantErrContains: "exceeds maximum length",
		},
		{
			name:            "dangerous chars in product",
			product:         "pkg:npm/lodash@4.17.21;malicious",
			vulnerability:   "CVE-2023-1234",
			status:          "not_affected",
			justification:   "component_not_present",
			wantErrContains: "dangerous characters",
		},
		{
			name:            "invalid status",
			product:         "pkg:npm/lodash@4.17.21",
			vulnerability:   "CVE-2023-1234",
			status:          "invalid_status",
			wantErrContains: "invalid status",
		},
		{
			name:            "invalid justification",
			product:         "pkg:npm/lodash@4.17.21",
			vulnerability:   "CVE-2023-1234",
			status:          "not_affected",
			justification:   "invalid_justification",
			wantErrContains: "invalid justification",
		},
		{
			name:            "not_affected without justification or impact",
			product:         "pkg:npm/lodash@4.17.21",
			vulnerability:   "CVE-2023-1234",
			status:          "not_affected",
			wantErrContains: "either justification or impact statement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.CreateStatement(
				tt.product,
				tt.vulnerability,
				tt.status,
				tt.justification,
				tt.impactStatement,
				tt.actionStatement,
				tt.author,
			)

			if err == nil {
				t.Fatal("CreateStatement() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Errorf("CreateStatement() error = %v, want to contain %v", err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestMergeDocuments_Success(t *testing.T) {
	client := NewClient("test-author")

	// Create two test documents
	doc1JSON := `{
		"@context": "https://openvex.dev/ns",
		"@id": "doc1",
		"author": "author1",
		"version": 1,
		"timestamp": "2023-01-01T00:00:00Z",
		"statements": [
			{
				"vulnerability": {"name": "CVE-2023-1234"},
				"products": [{"@id": "pkg:npm/lodash@4.17.21"}],
				"status": "not_affected",
				"justification": "component_not_present"
			}
		]
	}`

	doc2JSON := `{
		"@context": "https://openvex.dev/ns",
		"@id": "doc2",
		"author": "author2",
		"version": 1,
		"timestamp": "2023-01-02T00:00:00Z",
		"statements": [
			{
				"vulnerability": {"name": "CVE-2023-5678"},
				"products": [{"@id": "pkg:npm/express@4.18.0"}],
				"status": "affected"
			}
		]
	}`

	var doc1Map, doc2Map map[string]interface{}
	json.Unmarshal([]byte(doc1JSON), &doc1Map)
	json.Unmarshal([]byte(doc2JSON), &doc2Map)

	input := &MergeInput{
		Documents: []map[string]interface{}{doc1Map, doc2Map},
		Author:    "merger",
		ID:        "merged-doc",
	}

	merged, err := client.MergeDocuments(input)
	if err != nil {
		t.Fatalf("MergeDocuments() error = %v", err)
	}

	// Verify merged document
	if merged.ID != "merged-doc" {
		t.Errorf("ID = %v, want merged-doc", merged.ID)
	}
	if merged.Author != "merger" {
		t.Errorf("Author = %v, want merger", merged.Author)
	}
	if len(merged.Statements) != 2 {
		t.Errorf("Statements length = %v, want 2", len(merged.Statements))
	}
}

func TestMergeDocuments_WithFilters(t *testing.T) {
	client := NewClient("test-author")

	// Create document with multiple statements
	docJSON := `{
		"@context": "https://openvex.dev/ns",
		"@id": "doc1",
		"author": "author1",
		"version": 1,
		"timestamp": "2023-01-01T00:00:00Z",
		"statements": [
			{
				"vulnerability": {"name": "CVE-2023-1234"},
				"products": [{"@id": "pkg:npm/lodash@4.17.21"}],
				"status": "not_affected",
				"justification": "component_not_present"
			},
			{
				"vulnerability": {"name": "CVE-2023-5678"},
				"products": [{"@id": "pkg:npm/express@4.18.0"}],
				"status": "affected"
			}
		]
	}`

	var docMap map[string]interface{}
	json.Unmarshal([]byte(docJSON), &docMap)

	t.Run("filter by product", func(t *testing.T) {
		input := &MergeInput{
			Documents: []map[string]interface{}{docMap, docMap},
			Products:  []string{"pkg:npm/lodash@4.17.21"},
		}

		merged, err := client.MergeDocuments(input)
		if err != nil {
			t.Fatalf("MergeDocuments() error = %v", err)
		}

		// Should only have lodash statements
		if len(merged.Statements) == 0 {
			t.Error("Expected filtered statements, got none")
		}
		for _, stmt := range merged.Statements {
			if stmt.Products[0].Component.ID != "pkg:npm/lodash@4.17.21" {
				t.Errorf("Unexpected product in filtered result: %v", stmt.Products[0].Component.ID)
			}
		}
	})

	t.Run("filter by vulnerability", func(t *testing.T) {
		input := &MergeInput{
			Documents:       []map[string]interface{}{docMap, docMap},
			Vulnerabilities: []string{"CVE-2023-1234"},
		}

		merged, err := client.MergeDocuments(input)
		if err != nil {
			t.Fatalf("MergeDocuments() error = %v", err)
		}

		// Should only have CVE-2023-1234 statements
		if len(merged.Statements) == 0 {
			t.Error("Expected filtered statements, got none")
		}
		for _, stmt := range merged.Statements {
			if string(stmt.Vulnerability.Name) != "CVE-2023-1234" {
				t.Errorf("Unexpected vulnerability in filtered result: %v", stmt.Vulnerability.Name)
			}
		}
	})
}

func TestMergeDocuments_ValidationErrors(t *testing.T) {
	client := NewClient("test-author")

	tests := []struct {
		name            string
		input           *MergeInput
		wantErrContains string
	}{
		{
			name: "too few documents",
			input: &MergeInput{
				Documents: []map[string]interface{}{
					{"@context": "https://openvex.dev/ns", "statements": []interface{}{}},
				},
			},
			wantErrContains: "at least 2",
		},
		{
			name: "invalid document structure - missing context",
			input: &MergeInput{
				Documents: []map[string]interface{}{
					{"statements": []interface{}{}},
					{"statements": []interface{}{}},
				},
			},
			wantErrContains: "@context",
		},
		{
			name: "invalid document structure - missing statements",
			input: &MergeInput{
				Documents: []map[string]interface{}{
					{"@context": "https://openvex.dev/ns"},
					{"@context": "https://openvex.dev/ns"},
				},
			},
			wantErrContains: "statements",
		},
		{
			name: "author too long",
			input: &MergeInput{
				Documents: []map[string]interface{}{
					{"@context": "https://openvex.dev/ns", "statements": []interface{}{}},
					{"@context": "https://openvex.dev/ns", "statements": []interface{}{}},
				},
				Author: strings.Repeat("a", 201),
			},
			wantErrContains: "exceeds maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.MergeDocuments(tt.input)
			if err == nil {
				t.Fatal("MergeDocuments() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Errorf("MergeDocuments() error = %v, want to contain %v", err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		wantStatus vexlib.Status
		wantErr    bool
	}{
		{
			name:       "not_affected",
			status:     "not_affected",
			wantStatus: vexlib.StatusNotAffected,
			wantErr:    false,
		},
		{
			name:       "affected",
			status:     "affected",
			wantStatus: vexlib.StatusAffected,
			wantErr:    false,
		},
		{
			name:       "fixed",
			status:     "fixed",
			wantStatus: vexlib.StatusFixed,
			wantErr:    false,
		},
		{
			name:       "under_investigation",
			status:     "under_investigation",
			wantStatus: vexlib.StatusUnderInvestigation,
			wantErr:    false,
		},
		{
			name:    "invalid status",
			status:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStatus(tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantStatus {
				t.Errorf("parseStatus() = %v, want %v", got, tt.wantStatus)
			}
		})
	}
}

func TestParseJustification(t *testing.T) {
	tests := []struct {
		name              string
		justification     string
		wantJustification vexlib.Justification
		wantErr           bool
	}{
		{
			name:              "component_not_present",
			justification:     "component_not_present",
			wantJustification: vexlib.ComponentNotPresent,
			wantErr:           false,
		},
		{
			name:              "vulnerable_code_not_present",
			justification:     "vulnerable_code_not_present",
			wantJustification: vexlib.VulnerableCodeNotPresent,
			wantErr:           false,
		},
		{
			name:              "vulnerable_code_not_in_execute_path",
			justification:     "vulnerable_code_not_in_execute_path",
			wantJustification: vexlib.VulnerableCodeNotInExecutePath,
			wantErr:           false,
		},
		{
			name:              "vulnerable_code_cannot_be_controlled_by_adversary",
			justification:     "vulnerable_code_cannot_be_controlled_by_adversary",
			wantJustification: vexlib.VulnerableCodeCannotBeControlledByAdversary,
			wantErr:           false,
		},
		{
			name:              "inline_mitigations_already_exist",
			justification:     "inline_mitigations_already_exist",
			wantJustification: vexlib.InlineMitigationsAlreadyExist,
			wantErr:           false,
		},
		{
			name:          "invalid justification",
			justification: "invalid",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseJustification(tt.justification)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseJustification() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantJustification {
				t.Errorf("parseJustification() = %v, want %v", got, tt.wantJustification)
			}
		})
	}
}
