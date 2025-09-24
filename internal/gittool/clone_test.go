package gittool

import (
	"os"
	"testing"
)

func TestAdaptUrlIfTokenIsProvided(t *testing.T) {
	origToken := os.Getenv("GIT_TOKEN")
	defer os.Setenv("GIT_TOKEN", origToken)

	tests := []struct {
		name     string
		token    string
		repoURL  string
		expected string
	}{
		{
			name:     "No token, https URL",
			token:    "",
			repoURL:  "https://github.com/BjarneRentz/central-cyclone",
			expected: "https://github.com/BjarneRentz/central-cyclone",
		},
		{
			name:     "Token, https URL",
			token:    "mytoken",
			repoURL:  "https://github.com/BjarneRentz/central-cyclone",
			expected: "https://mytoken@github.com/BjarneRentz/central-cyclone",
		},
		{
			name:     "Token, https  Azure DevopsURL",
			token:    "mytoken",
			repoURL:  "https://dev.azure.com/my-org/my-repo",
			expected: "https://mytoken@dev.azure.com/my-org/my-repo",
		},
		{
			name:     "Token, non-https URL",
			token:    "mytoken",
			repoURL:  "git@github.com:BjarneRentz/central-cyclone.git",
			expected: "git@github.com:BjarneRentz/central-cyclone.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("GIT_TOKEN", tt.token)
			result := adaptUrlIfTokenIsProvided(tt.repoURL)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
