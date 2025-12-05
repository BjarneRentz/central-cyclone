package workspace

import (
	"fmt"
	"net/url"
	"strings"
)

// RepoURLMapper defines methods for mapping repository URLs to folder names
type RepoURLMapper interface {
	// GetFolderName converts a repository URL to a folder name
	GetFolderName(repoURL string) (string, error)
}

// DefaultRepoMapper implements the default URL to folder name mapping strategy
type DefaultRepoMapper struct{}

// GetFolderName converts repository URLs to folder names:
// - GitHub: org_repo
// - Azure DevOps: org_project_repo
func (m DefaultRepoMapper) GetFolderName(repoURL string) (string, error) {
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("invalid repo URL: %w", err)
	}

	path := strings.TrimSuffix(parsedURL.Path, ".git")
	pathParts := strings.Split(path, "/")

	// Remove leading slash if present
	if len(pathParts) > 0 && pathParts[0] == "" {
		pathParts = pathParts[1:]
	}

	switch parsedURL.Host {
	case "dev.azure.com":
		// Azure DevOps: /org/project/_git/repo
		for i, part := range pathParts {
			if part == "_git" && i > 0 && i < len(pathParts)-1 {
				org := pathParts[0]
				project := pathParts[1]
				repo := pathParts[i+1]
				return fmt.Sprintf("%s_%s_%s", org, project, repo), nil
			}
		}
		return "", fmt.Errorf("invalid Azure DevOps URL format: %s", repoURL)

	case "github.com":
		// GitHub: /org/repo
		if len(pathParts) >= 2 {
			org := pathParts[len(pathParts)-2]
			repo := pathParts[len(pathParts)-1]
			return fmt.Sprintf("%s_%s", org, repo), nil
		}
		return "", fmt.Errorf("invalid GitHub URL format: %s", repoURL)

	default:
		return "", fmt.Errorf("unsupported git host: %s", parsedURL.Host)
	}
}
