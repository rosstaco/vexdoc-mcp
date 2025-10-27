package vex

import (
	"strings"
	"testing"
)

func TestValidateStringLength(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     string
		maxLength int
		wantErr   bool
	}{
		{
			name:      "empty string is valid",
			fieldName: "test",
			value:     "",
			maxLength: 10,
			wantErr:   false,
		},
		{
			name:      "string within limit",
			fieldName: "test",
			value:     "hello",
			maxLength: 10,
			wantErr:   false,
		},
		{
			name:      "string at exact limit",
			fieldName: "test",
			value:     "1234567890",
			maxLength: 10,
			wantErr:   false,
		},
		{
			name:      "string exceeds limit",
			fieldName: "test",
			value:     "12345678901",
			maxLength: 10,
			wantErr:   true,
		},
		{
			name:      "very long string",
			fieldName: "product",
			value:     strings.Repeat("a", 1001),
			maxLength: MaxStringLength,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStringLength(tt.fieldName, tt.value, tt.maxLength)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStringLength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDangerousChars(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     string
		wantErr   bool
	}{
		{
			name:      "empty string is valid",
			fieldName: "test",
			value:     "",
			wantErr:   false,
		},
		{
			name:      "safe alphanumeric",
			fieldName: "test",
			value:     "hello123",
			wantErr:   false,
		},
		{
			name:      "safe with spaces and dashes",
			fieldName: "author",
			value:     "security-team@example.com",
			wantErr:   false,
		},
		{
			name:      "PURL format",
			fieldName: "product",
			value:     "pkg:npm/lodash@4.17.21",
			wantErr:   false,
		},
		{
			name:      "CVE format",
			fieldName: "vulnerability",
			value:     "CVE-2023-1234",
			wantErr:   false,
		},
		{
			name:      "dangerous semicolon",
			fieldName: "test",
			value:     "hello;world",
			wantErr:   true,
		},
		{
			name:      "dangerous pipe",
			fieldName: "test",
			value:     "hello|world",
			wantErr:   true,
		},
		{
			name:      "dangerous ampersand",
			fieldName: "test",
			value:     "hello&world",
			wantErr:   true,
		},
		{
			name:      "dangerous backtick",
			fieldName: "test",
			value:     "hello`world",
			wantErr:   true,
		},
		{
			name:      "dangerous dollar",
			fieldName: "test",
			value:     "hello$world",
			wantErr:   true,
		},
		{
			name:      "dangerous parentheses",
			fieldName: "test",
			value:     "hello(world)",
			wantErr:   true,
		},
		{
			name:      "dangerous brackets",
			fieldName: "test",
			value:     "hello[world]",
			wantErr:   true,
		},
		{
			name:      "dangerous braces",
			fieldName: "test",
			value:     "hello{world}",
			wantErr:   true,
		},
		{
			name:      "dangerous angle brackets",
			fieldName: "test",
			value:     "hello<world>",
			wantErr:   true,
		},
		{
			name:      "dangerous quotes",
			fieldName: "test",
			value:     "hello'world",
			wantErr:   true,
		},
		{
			name:      "dangerous backslash",
			fieldName: "test",
			value:     "hello\\world",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDangerousChars(tt.fieldName, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDangerousChars() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     string
		wantErr   bool
	}{
		{
			name:      "empty string fails",
			fieldName: "test",
			value:     "",
			wantErr:   true,
		},
		{
			name:      "non-empty string passes",
			fieldName: "test",
			value:     "value",
			wantErr:   false,
		},
		{
			name:      "whitespace only is valid (not our job to trim)",
			fieldName: "test",
			value:     "   ",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.fieldName, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequired() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDocumentCount(t *testing.T) {
	tests := []struct {
		name    string
		count   int
		wantErr bool
	}{
		{
			name:    "zero documents fails",
			count:   0,
			wantErr: true,
		},
		{
			name:    "one document fails (need at least 2)",
			count:   1,
			wantErr: true,
		},
		{
			name:    "two documents passes",
			count:   2,
			wantErr: false,
		},
		{
			name:    "max documents passes",
			count:   MaxMergeDocuments,
			wantErr: false,
		},
		{
			name:    "over max documents fails",
			count:   MaxMergeDocuments + 1,
			wantErr: true,
		},
		{
			name:    "many over max documents fails",
			count:   100,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDocumentCount(tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDocumentCount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationConstants(t *testing.T) {
	// Verify constants are set to reasonable values
	if MaxStringLength != 1000 {
		t.Errorf("MaxStringLength = %d, want 1000", MaxStringLength)
	}
	if MaxAuthorLength != 200 {
		t.Errorf("MaxAuthorLength = %d, want 200", MaxAuthorLength)
	}
	if MaxIDLength != 500 {
		t.Errorf("MaxIDLength = %d, want 500", MaxIDLength)
	}
	if MaxMergeDocuments != 20 {
		t.Errorf("MaxMergeDocuments = %d, want 20", MaxMergeDocuments)
	}
	if MinMergeDocuments != 2 {
		t.Errorf("MinMergeDocuments = %d, want 2", MinMergeDocuments)
	}
}
