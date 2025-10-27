package workspace

import (
	"central-cyclone/internal/gittool"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const workspacePath = "workfolder"
const repoFolder = "repos"
const sbomFolder = "sboms"

type localWorkspace struct {
	path      string
	reposPath string
	sbomsPath string
	gitCloner gittool.Cloner
}

type Workspace interface {
	Clear() error
	CloneRepoToWorkspace(repoUrl string) (string, error)
}

func (w localWorkspace) CloneRepoToWorkspace(repoUrl string) (string, error) {
	folderName, err := getFolderNameForRepoUrl(repoUrl)
	if err != nil {
		return "", fmt.Errorf("failed to get folder name from repo URL: %w", err)
	}
	targetDir := filepath.Join(w.reposPath, folderName)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create target dir: %w", err)
		}
	}

	err = w.gitCloner.CloneRepoToDir(repoUrl, targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to clone repo: %w", err)
	}
	return targetDir, nil
}

func (w localWorkspace) Clear() error {

	if _, err := os.Stat(w.reposPath); !os.IsNotExist(err) {
		repos, err := os.ReadDir(w.reposPath)
		if err != nil {
			return fmt.Errorf("failed to read repos directory: %w", err)
		}
		for _, entry := range repos {
			entryPath := filepath.Join(w.reposPath, entry.Name())
			if err := os.RemoveAll(entryPath); err != nil {
				return fmt.Errorf("failed to remove '%s': %w", entryPath, err)
			}
		}
	}

	if _, err := os.Stat(w.sbomsPath); !os.IsNotExist(err) {
		sboms, err := os.ReadDir(w.sbomsPath)
		if err != nil {
			return fmt.Errorf("failed to read sboms directory: %w", err)
		}
		for _, entry := range sboms {
			entryPath := filepath.Join(w.sbomsPath, entry.Name())
			if err := os.RemoveAll(entryPath); err != nil {
				return fmt.Errorf("failed to remove '%s': %w", entryPath, err)
			}
		}
	}
	return nil
}

func (w localWorkspace) List() ([]string, error) {
	entries, err := os.ReadDir(w.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name())
	}
	return files, nil
}

func CreateLocalWorkspace() (Workspace, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	fullWorkFolderPath := filepath.Join(homeDir, ".central-cyclone", workspacePath)

	err = createFolderIfNotExists(fullWorkFolderPath)
	if err != nil {
		return nil, fmt.Errorf("could not create workfolder: %w", err)
	}

	fullReposPath := filepath.Join(fullWorkFolderPath, repoFolder)
	err = createFolderIfNotExists(fullReposPath)
	if err != nil {
		return nil, fmt.Errorf("could not create workfolder: %w", err)
	}

	fullSbomsPath := filepath.Join(fullWorkFolderPath, sbomFolder)
	err = createFolderIfNotExists(fullSbomsPath)
	if err != nil {
		return nil, fmt.Errorf("could not create workfolder: %w", err)
	}
	return localWorkspace{
		path:      fullWorkFolderPath,
		reposPath: fullReposPath,
		sbomsPath: fullSbomsPath,
		gitCloner: gittool.LocalGitCloner{},
	}, nil
}

func createFolderIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create folder '%s': %w", path, err)
		}
	}
	return nil
}

func getFolderNameForRepoUrl(repoUrl string) (string, error) {
	parsedUrl, err := url.Parse(repoUrl)
	if err != nil {
		return "", fmt.Errorf("invalid repo URL: %w", err)
	}

	path := strings.TrimSuffix(parsedUrl.Path, ".git")
	pathParts := strings.Split(path, "/")

	// Remove leading slash if present
	if len(pathParts) > 0 && pathParts[0] == "" {
		pathParts = pathParts[1:]
	}

	switch parsedUrl.Host {
	case "dev.azure.com":
		// Azure DevOps: /org/project/_git/repo
		for i, part := range pathParts {
			if part == "_git" && i > 0 && i < len(pathParts)-1 {
				org := pathParts[0]
				project := pathParts[1]
				repo := pathParts[i+1]
				folderName := fmt.Sprintf("%s_%s_%s", org, project, repo)
				return folderName, nil
			}
		}
	default:
		// GitHub: /org/repo
		if len(pathParts) >= 2 {
			org := pathParts[len(pathParts)-2]
			repo := pathParts[len(pathParts)-1]
			folderName := fmt.Sprintf("%s_%s", org, repo)
			return folderName, nil
		}
	}

	return "", fmt.Errorf("unexpected repo URL format: %s", repoUrl)
}
