package vex

import (
	"encoding/json"
	"fmt"
	"time"

	vexlib "github.com/openvex/go-vex/pkg/vex"
)

// Client handles VEX operations using the native go-vex library
type Client struct {
	defaultAuthor string
}

// NewClient creates a new VEX client
func NewClient(defaultAuthor string) *Client {
	if defaultAuthor == "" {
		defaultAuthor = "vexdoc-mcp-server"
	}
	return &Client{
		defaultAuthor: defaultAuthor,
	}
}

// CreateInput represents the input for creating a VEX statement
type CreateInput struct {
	Product         string
	Vulnerability   string
	Status          string
	Justification   string
	ImpactStatement string
	ActionStatement string
	Author          string
}

// MergeInput represents the input for merging VEX documents
type MergeInput struct {
	Documents       []map[string]interface{}
	Author          string
	AuthorRole      string
	ID              string
	Products        []string
	Vulnerabilities []string
}

// CreateStatement creates a new VEX statement following the vexctl pattern
func (c *Client) CreateStatement(
	product string,
	vulnerability string,
	status string,
	justification string,
	impactStatement string,
	actionStatement string,
	author string,
) (*vexlib.VEX, error) {
	// Security boundary checks (DoS prevention, defense in depth)
	if err := ValidateRequired("product", product); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	if err := ValidateStringLength("product", product, MaxStringLength); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	if err := ValidateDangerousChars("product", product); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	if err := ValidateRequired("vulnerability", vulnerability); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	if err := ValidateStringLength("vulnerability", vulnerability, MaxStringLength); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	if err := ValidateRequired("status", status); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Optional fields - only check length/chars if provided
	if err := ValidateStringLength("justification", justification, MaxStringLength); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	if err := ValidateStringLength("impact_statement", impactStatement, MaxStringLength); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	if err := ValidateDangerousChars("impact_statement", impactStatement); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	if err := ValidateStringLength("action_statement", actionStatement, MaxStringLength); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	if err := ValidateDangerousChars("action_statement", actionStatement); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	if err := ValidateStringLength("author", author, MaxAuthorLength); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	if err := ValidateDangerousChars("author", author); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Create new VEX document
	doc := vexlib.New()
	now := time.Now()

	// Set metadata
	doc.Context = vexlib.Context
	doc.ID = fmt.Sprintf("vex-%d", now.Unix())
	doc.Author = c.getAuthor(author)
	doc.Version = 1
	doc.Timestamp = &now

	// Parse status - let go-vex handle invalid values
	vexStatus, err := parseStatus(status)
	if err != nil {
		return nil, err
	}

	// Create statement
	statement := vexlib.Statement{
		Vulnerability: vexlib.Vulnerability{
			Name: vexlib.VulnerabilityID(vulnerability),
		},
		Products: []vexlib.Product{
			{
				Component: vexlib.Component{
					ID: product,
				},
			},
		},
		Status: vexStatus,
	}

	// Add justification if provided (for not_affected status)
	if justification != "" {
		just, err := parseJustification(justification)
		if err != nil {
			return nil, err
		}
		statement.Justification = just
	}

	// Add impact statement if provided
	if impactStatement != "" {
		statement.ImpactStatement = impactStatement
	}

	// Add action statement if provided
	if actionStatement != "" {
		statement.ActionStatement = actionStatement
	}

	// Add statement to document
	doc.Statements = append(doc.Statements, statement)

	// Let go-vex validate the statement (domain validation)
	if err := statement.Validate(); err != nil {
		return nil, fmt.Errorf("statement validation failed: %w", err)
	}

	return &doc, nil
}

// MergeDocuments merges multiple VEX documents using the native library
func (c *Client) MergeDocuments(input *MergeInput) (*vexlib.VEX, error) {
	// Security boundary checks
	if err := ValidateDocumentCount(len(input.Documents)); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	if err := ValidateStringLength("author", input.Author, MaxAuthorLength); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	if err := ValidateDangerousChars("author", input.Author); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	if err := ValidateStringLength("author_role", input.AuthorRole, MaxAuthorLength); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	if err := ValidateDangerousChars("author_role", input.AuthorRole); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	if err := ValidateStringLength("id", input.ID, MaxIDLength); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	if err := ValidateDangerousChars("id", input.ID); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Validate products list
	for i, product := range input.Products {
		if err := ValidateStringLength(fmt.Sprintf("products[%d]", i), product, MaxStringLength); err != nil {
			return nil, fmt.Errorf("validation error: %w", err)
		}
		if err := ValidateDangerousChars(fmt.Sprintf("products[%d]", i), product); err != nil {
			return nil, fmt.Errorf("validation error: %w", err)
		}
	}

	// Validate vulnerabilities list
	for i, vuln := range input.Vulnerabilities {
		if err := ValidateStringLength(fmt.Sprintf("vulnerabilities[%d]", i), vuln, MaxStringLength); err != nil {
			return nil, fmt.Errorf("validation error: %w", err)
		}
	}

	// Validate each document has basic structure
	for i, doc := range input.Documents {
		if _, hasContext := doc["@context"]; !hasContext {
			return nil, fmt.Errorf("document %d must be a valid VEX document with @context", i+1)
		}
		if _, hasStatements := doc["statements"]; !hasStatements {
			return nil, fmt.Errorf("document %d must be a valid VEX document with statements", i+1)
		}
	}

	// Parse documents from JSON
	var docs []*vexlib.VEX
	for i, docData := range input.Documents {
		// Convert map to JSON bytes
		jsonBytes, err := json.Marshal(docData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal document %d: %w", i+1, err)
		}

		// Parse VEX document - let go-vex validate the structure
		doc, err := vexlib.Parse(jsonBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse document %d: %w", i+1, err)
		}

		docs = append(docs, doc)
	}

	// Merge documents using the library
	merged, err := vexlib.MergeDocuments(docs)
	if err != nil {
		return nil, fmt.Errorf("failed to merge documents: %w", err)
	}

	// Apply custom metadata if provided
	if input.ID != "" {
		merged.ID = input.ID
	}
	if input.Author != "" {
		merged.Author = input.Author
	}
	if input.AuthorRole != "" {
		merged.AuthorRole = input.AuthorRole
	}

	// Filter by products if specified
	if len(input.Products) > 0 {
		merged = c.filterByProducts(merged, input.Products)
	}

	// Filter by vulnerabilities if specified
	if len(input.Vulnerabilities) > 0 {
		merged = c.filterByVulnerabilities(merged, input.Vulnerabilities)
	}

	// Update timestamp
	now := time.Now()
	merged.Timestamp = &now

	return merged, nil
}

// getAuthor returns the author or default
func (c *Client) getAuthor(author string) string {
	if author != "" {
		return author
	}
	return c.defaultAuthor
}

// filterByProducts filters statements to only include specified products
func (c *Client) filterByProducts(doc *vexlib.VEX, products []string) *vexlib.VEX {
	var filtered []vexlib.Statement
	productSet := make(map[string]bool)
	for _, p := range products {
		productSet[p] = true
	}

	for _, stmt := range doc.Statements {
		for _, prod := range stmt.Products {
			if productSet[prod.Component.ID] {
				filtered = append(filtered, stmt)
				break
			}
		}
	}

	doc.Statements = filtered
	return doc
}

// filterByVulnerabilities filters statements to only include specified vulnerabilities
func (c *Client) filterByVulnerabilities(doc *vexlib.VEX, vulnerabilities []string) *vexlib.VEX {
	var filtered []vexlib.Statement
	vulnSet := make(map[string]bool)
	for _, v := range vulnerabilities {
		vulnSet[v] = true
	}

	for _, stmt := range doc.Statements {
		if vulnSet[string(stmt.Vulnerability.Name)] {
			filtered = append(filtered, stmt)
		}
	}

	doc.Statements = filtered
	return doc
}

// parseStatus converts string status to vex.Status
func parseStatus(status string) (vexlib.Status, error) {
	switch status {
	case "not_affected":
		return vexlib.StatusNotAffected, nil
	case "affected":
		return vexlib.StatusAffected, nil
	case "fixed":
		return vexlib.StatusFixed, nil
	case "under_investigation":
		return vexlib.StatusUnderInvestigation, nil
	default:
		return "", fmt.Errorf("invalid status: %s", status)
	}
}

// parseJustification converts string justification to vex.Justification
func parseJustification(justification string) (vexlib.Justification, error) {
	switch justification {
	case "component_not_present":
		return vexlib.ComponentNotPresent, nil
	case "vulnerable_code_not_present":
		return vexlib.VulnerableCodeNotPresent, nil
	case "vulnerable_code_not_in_execute_path":
		return vexlib.VulnerableCodeNotInExecutePath, nil
	case "vulnerable_code_cannot_be_controlled_by_adversary":
		return vexlib.VulnerableCodeCannotBeControlledByAdversary, nil
	case "inline_mitigations_already_exist":
		return vexlib.InlineMitigationsAlreadyExist, nil
	default:
		return "", fmt.Errorf("invalid justification: %s", justification)
	}
}
