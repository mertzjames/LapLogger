package handlers

import (
	"testing"
)

func TestSlugRegex_MatchesNonAlphanumeric(t *testing.T) {
	tests := []struct {
		char    string
		matches bool
	}{
		{"a", false},
		{"z", false},
		{"0", false},
		{"9", false},
		{" ", true},
		{"!", true},
		{"@", true},
		{"-", true},
		{"_", true},
		{"A", true},
	}

	for _, tt := range tests {
		t.Run(tt.char, func(t *testing.T) {
			got := slugRe.MatchString(tt.char)
			if got != tt.matches {
				t.Errorf("slugRe.MatchString(%q) = %v; want %v", tt.char, got, tt.matches)
			}
		})
	}
}

func TestNewLeagueHandler(t *testing.T) {
	h := NewLeagueHandler(nil, nil)
	if h == nil {
		t.Fatal("NewLeagueHandler returned nil")
	}
}
