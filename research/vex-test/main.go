// main.go - Basic VEX API test
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/openvex/go-vex/pkg/vex"
)

func main() {
	fmt.Println("Testing VEX Go Library APIs...")

	// Test 1: Create basic VEX document using New() function
	doc := vex.New()

	// Set metadata
	now := time.Now()
	doc.Context = "https://openvex.dev/ns/v0.2.5"
	doc.ID = "test-vex-001"
	doc.Author = "research@example.com"
	doc.Version = 1
	doc.Timestamp = &now

	// Test 2: Create VEX statement
	statement := vex.Statement{
		Vulnerability: vex.Vulnerability{Name: "CVE-2023-1234"},
		Products: []vex.Product{{
			Component: vex.Component{ID: "test-product"},
		}},
		Status:        vex.StatusNotAffected,
		Justification: vex.ComponentNotPresent,
	}

	doc.Statements = append(doc.Statements, statement)

	// Test 3: Serialize to JSON
	jsonBytes, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		log.Fatal("Failed to marshal VEX document:", err)
	}

	fmt.Printf("Generated VEX document:\n%s\n", string(jsonBytes))

	// Test 4: Document API exploration
	testDocumentAPIs(&doc)
}

func testDocumentAPIs(doc *vex.VEX) {
	fmt.Println("\n=== API Exploration ===")

	// Test statement validation
	if len(doc.Statements) > 0 {
		if err := doc.Statements[0].Validate(); err != nil {
			fmt.Printf("Statement validation failed: %v\n", err)
		} else {
			fmt.Println("✓ Statement validation passed")
		}
	}

	// Test statement manipulation
	fmt.Printf("Document has %d statements\n", len(doc.Statements))

	// Document available methods
	fmt.Println("\nAvailable VEX operations to document:")
	fmt.Println("- Document creation: ✓")
	fmt.Println("- Statement creation: ✓")
	fmt.Println("- JSON serialization: ✓")
	fmt.Println("- Statement validation: ✓")
	fmt.Println("- Merge operations: ?") // To be tested
	fmt.Println("- Streaming: ?")        // To be tested

	// Test additional operations
	testMergeOperations()
	testFileOperations()
}

func testMergeOperations() {
	fmt.Println("\n=== Testing Merge Operations ===")

	// Create second document for merging
	doc2 := vex.New()
	now := time.Now()
	doc2.Context = "https://openvex.dev/ns/v0.2.5"
	doc2.ID = "test-vex-002"
	doc2.Author = "research@example.com"
	doc2.Version = 1
	doc2.Timestamp = &now

	statement2 := vex.Statement{
		Vulnerability: vex.Vulnerability{Name: "CVE-2023-5678"},
		Products: []vex.Product{{
			Component: vex.Component{ID: "test-product-2"},
		}},
		Status:        vex.StatusAffected,
		Justification: "",
	}

	doc2.Statements = append(doc2.Statements, statement2)

	fmt.Printf("Created second document with %d statements\n", len(doc2.Statements))
	fmt.Println("✓ Multiple document creation: ✓")

	// Note: Actual merge functionality needs to be explored in the vex package
	fmt.Println("- Merge operations: Need to explore vex package APIs")
}

func testFileOperations() {
	fmt.Println("\n=== Testing File Operations ===")

	// Test writing to file
	doc := vex.New()
	now := time.Now()
	doc.Context = "https://openvex.dev/ns/v0.2.5"
	doc.ID = "test-file-vex"
	doc.Author = "research@example.com"
	doc.Version = 1
	doc.Timestamp = &now

	statement := vex.Statement{
		Vulnerability: vex.Vulnerability{Name: "CVE-2023-9999"},
		Products: []vex.Product{{
			Component: vex.Component{ID: "file-test-product"},
		}},
		Status:        vex.StatusUnderInvestigation,
		Justification: "",
	}

	doc.Statements = append(doc.Statements, statement)

	// Write to temporary file
	jsonBytes, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal document: %v\n", err)
		return
	}

	// For now, just confirm we can serialize
	fmt.Printf("✓ Document serialization for file output: %d bytes\n", len(jsonBytes))
	fmt.Println("- File I/O operations: Available via standard JSON marshaling")

	// Test merge functionality
	testMergeWithExistingDoc(&doc)
}

func testMergeWithExistingDoc(doc1 *vex.VEX) {
	fmt.Println("\n=== Testing Merge Functionality ===")

	// Create another document
	doc2 := vex.New()
	now := time.Now()
	doc2.Context = "https://openvex.dev/ns/v0.2.5"
	doc2.ID = "test-merge-doc"
	doc2.Author = "research@example.com"
	doc2.Version = 1
	doc2.Timestamp = &now

	statement := vex.Statement{
		Vulnerability: vex.Vulnerability{Name: "CVE-2023-0001"},
		Products: []vex.Product{{
			Component: vex.Component{ID: "merge-test-product"},
		}},
		Status:        vex.StatusNotAffected,
		Justification: vex.ComponentNotPresent,
	}

	doc2.Statements = append(doc2.Statements, statement)

	// Test the merge function
	fmt.Printf("Document 1 has %d statements\n", len(doc1.Statements))
	fmt.Printf("Document 2 has %d statements\n", len(doc2.Statements))

	// Test MergeDocuments function
	merged, err := vex.MergeDocuments([]*vex.VEX{doc1, &doc2})
	if err != nil {
		fmt.Printf("Merge failed: %v\n", err)
	} else {
		fmt.Printf("✓ Merge successful! Merged document has %d statements\n", len(merged.Statements))
	}
}
