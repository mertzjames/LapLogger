package handlers

import (
	"strings"
	"testing"
)

// === SEC-1: League Input Validation Security Tests ===

// TestSecurity_Slug_XSSCharacters verifies that the slug generation
// properly strips HTML/JS characters that could be used for stored XSS.
func TestSecurity_Slug_XSSCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string // slug should NOT contain these
	}{
		{"script tag", "<script>alert(1)</script>", "<"},
		{"html entities", "League&amp;Name", "&"},
		{"angle brackets", "League<>Name", "<"},
		{"double quotes", `League"Name`, `"`},
		{"single quotes", "League'Name", "'"},
		{"event handler", `League" onmouseover="alert(1)`, `"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slug := slugRe.ReplaceAllString(strings.ToLower(tt.input), "-")
			slug = strings.Trim(slug, "-")
			if strings.Contains(slug, tt.contains) {
				t.Errorf("slug %q still contains %q from input %q", slug, tt.contains, tt.input)
			}
		})
	}
}

// TestSecurity_Slug_SQLInjection verifies that SQL injection characters
// are stripped from generated slugs.
func TestSecurity_Slug_SQLInjection(t *testing.T) {
	inputs := []string{
		"'; DROP TABLE leagues;--",
		"name OR 1=1",
		"UNION SELECT * FROM users",
		"name; DELETE FROM leagues",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			slug := slugRe.ReplaceAllString(strings.ToLower(input), "-")
			slug = strings.Trim(slug, "-")
			// Slug should only contain [a-z0-9-]
			for _, c := range slug {
				if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
					t.Errorf("slug %q contains invalid char %q from input %q", slug, string(c), input)
				}
			}
		})
	}
}

// TestSecurity_Slug_PathTraversal verifies that path traversal sequences
// are neutralized in slug generation.
func TestSecurity_Slug_PathTraversal(t *testing.T) {
	inputs := []string{
		"../../../etc/passwd",
		"league/../../admin",
		"..\\..\\windows\\system32",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			slug := slugRe.ReplaceAllString(strings.ToLower(input), "-")
			slug = strings.Trim(slug, "-")
			if strings.Contains(slug, "..") {
				t.Errorf("slug %q contains path traversal from input %q", slug, input)
			}
			if strings.Contains(slug, "/") {
				t.Errorf("slug %q contains slash from input %q", slug, input)
			}
		})
	}
}

// TestSecurity_DBName_Generation verifies the DB name derived from league ID
// is safe for use in CREATE DATABASE statements.
func TestSecurity_DBName_Generation(t *testing.T) {
	// Simulate the pattern from CreateLeague handler:
	// dbName := fmt.Sprintf("laplogger_league_%s", strings.ReplaceAll(league.ID, "-", "_"))
	// The ID comes from uuid.New().String() which is [0-9a-f-]{36}

	sampleIDs := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
		"00000000-0000-0000-0000-000000000000",
	}

	for _, id := range sampleIDs {
		dbName := "laplogger_league_" + strings.ReplaceAll(id, "-", "_")
		// Must match the validDBName regex from provisioner.go
		if !isValidDBName(dbName) {
			t.Errorf("generated DB name %q is not valid", dbName)
		}
		// Must not contain SQL injection chars
		for _, c := range dbName {
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
				t.Errorf("DB name %q contains invalid char %q", dbName, string(c))
			}
		}
	}
}

// isValidDBName mirrors the regex from provisioner.go for test verification.
func isValidDBName(name string) bool {
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return len(name) > 0
}
