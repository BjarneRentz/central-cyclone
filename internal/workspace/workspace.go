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

type localWorkspace struct {
	path      string
	gitCloner gittool.Cloner
}

type Workspace interface {
	Clear() error
	List() ([]string, error)
	CloneRepoToWorkspace(repoUrl string) (string, error)
}

func (w localWorkspace) CloneRepoToWorkspace(repoUrl string) (string, error) {
	parsedUrl, err := url.Parse(repoUrl)
	if err != nil {
		return "", fmt.Errorf("invalid repo URL: %w", err)
	}

	// Example: https://github.com/org/repo.git -> org_repo
	pathParts := strings.Split(strings.TrimSuffix(parsedUrl.Path, ".git"), "/")
	if len(pathParts) < 3 {
		return "", fmt.Errorf("unexpected repo URL format: %s", repoUrl)
	}
	org := pathParts[1]
	repo := pathParts[2]
	folderName := fmt.Sprintf("%s_%s", org, repo)
	targetDir := filepath.Join(w.path, folderName)
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

// Removes all files and folders in the workspace directory
func (w localWorkspace) Clear() error {
	entries, err := os.ReadDir(w.path)
	if err != nil {
		return fmt.Errorf("failed to read workspace directory: %w", err)
	}
	for _, entry := range entries {
		entryPath := filepath.Join(w.path, entry.Name())
		if err := os.RemoveAll(entryPath); err != nil {
			return fmt.Errorf("failed to remove '%s': %w", entryPath, err)
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

	if _, err := os.Stat(fullWorkFolderPath); os.IsNotExist(err) {
		fmt.Printf("Creating work directory: %s\n", fullWorkFolderPath)
		if err := os.MkdirAll(fullWorkFolderPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create work folder '%s': %w", fullWorkFolderPath, err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to check work folder '%s': %w", fullWorkFolderPath, err)
	} else {
		fmt.Printf("Work directory '%s' already exists.\n", fullWorkFolderPath)
	}

	return localWorkspace{
		path:      fullWorkFolderPath,
		gitCloner: gittool.LocalGitCloner{},
	}, nil
}
