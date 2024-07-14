package cmd

import (
	"testing"
)

func TestNoDuplicateSuggestions(t *testing.T) {
	testCases := []struct {
		name        string
		suggestions []string
	}{
		{
			name:        "api_auditor_suggestions",
			suggestions: api_auditor_suggestions,
		},
		{
			name:        "cache_suggestions",
			suggestions: cache_suggestions,
		},
		{
			name:        "simple_suggestions",
			suggestions: proxyRun_suggestions,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use a map to track duplicates
			seen := make(map[string]bool)
			for _, suggestion := range tc.suggestions {
				if seen[suggestion] {
					t.Errorf("duplicate suggestion in %s: %s", tc.name, suggestion)
				}
				seen[suggestion] = true
			}
		})
	}

	t.Run("dupe check", func(t *testing.T) {
		seen := make(map[string]bool)
		for i, tc := range testCases {
			for _, suggestion := range tc.suggestions {
				if seen[suggestion] {
					t.Errorf("duplicate suggestion found: %s", suggestion)
				}
				seen[suggestion] = true
			}
			if i == len(testCases)-1 {
				// Reset seen for the next iteration
				seen = make(map[string]bool)
			}
		}
	})
}
