package vex

import (
	"fmt"
	"regexp"
)

// Security limits for DoS prevention
const (
	MaxStringLength   = 1000 // General max for most string fields
	MaxAuthorLength   = 200  // Shorter limit for author fields
	MaxIDLength       = 500  // Limit for custom IDs
	MaxMergeDocuments = 20   // Maximum documents to merge at once
	MinMergeDocuments = 2    // Minimum documents needed for merge
)

// Dangerous characters that could be used for injection attacks
// Defense in depth - even though we use native library, not subprocesses
var dangerousChars = regexp.MustCompile(`[;&|` + "`" + `$(){}[\]<>'"\\]`)

// ValidateStringLength checks if a string exceeds maximum length (DoS prevention)
func ValidateStringLength(name, value string, maxLength int) error {
	if value == "" {
		return nil // Empty is okay, let go-vex handle required field validation
	}
	if len(value) > maxLength {
		return fmt.Errorf("%s exceeds maximum length of %d characters", name, maxLength)
	}
	return nil
}

// ValidateDangerousChars checks for potentially dangerous characters (defense in depth)
func ValidateDangerousChars(name, value string) error {
	if value == "" {
		return nil
	}
	if dangerousChars.MatchString(value) {
		return fmt.Errorf("%s contains potentially dangerous characters", name)
	}
	return nil
}

// ValidateRequired checks if a required field is present
func ValidateRequired(name, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

// ValidateDocumentCount validates the number of documents for merging
func ValidateDocumentCount(count int) error {
	if count < MinMergeDocuments {
		return fmt.Errorf("at least %d VEX documents are required for merging", MinMergeDocuments)
	}
	if count > MaxMergeDocuments {
		return fmt.Errorf("maximum of %d documents can be merged at once", MaxMergeDocuments)
	}
	return nil
}
