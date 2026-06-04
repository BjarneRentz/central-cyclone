package gittool

import (
	"central-cyclone/internal/workspace"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
)

type Cloner interface {
	CloneRepo(repoURL string) (ClonedRepo, error)
	CloneOrUpdateRepo(repoURL string) (ClonedRepo, error)
}

func CreateLocalGitCloner(workspaceHandler workspace.Workspace) Cloner {
	return LocalGitCloner{workspace: workspaceHandler}
}

type LocalGitCloner struct {
	workspace workspace.Workspace
}

func (c LocalGitCloner) CloneRepo(repoURL string) (ClonedRepo, error) {
	path, err := c.workspace.CreateRepoFolder(repoURL)
	if err != nil {
		return ClonedRepo{}, err
	}

	slog.Info("🛠️  Cloning repo into the workfolder", "repo", repoURL)

	cloneOpts := &git.CloneOptions{
		URL:   repoURL,
		Depth: 1,
	}

	// Handle authentication if GIT_TOKEN is provided
	token := os.Getenv("GIT_TOKEN")
	if token != "" {
		cloneOpts.Auth = &http.BasicAuth{
			Username: "git",
			Password: token,
		}
	}

	_, err = git.PlainClone(path, cloneOpts)
	if err != nil {
		return ClonedRepo{}, err
	}

	return ClonedRepo{
		RepoUrl: repoURL,
		Path:    path,
	}, nil
}

func (c LocalGitCloner) CloneOrUpdateRepo(repoURL string) (ClonedRepo, error) {
	path, err := c.workspace.CreateRepoFolder(repoURL)
	if err != nil {
		return ClonedRepo{}, err
	}

	// Check if repository is already cloned
	if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
		// Repository exists, open and pull
		slog.Info("📁 Repository already exists, updating", "repo", repoURL)
		repo, err := git.PlainOpen(path)
		if err != nil {
			return ClonedRepo{}, fmt.Errorf("failed to open existing repository: %w", err)
		}

		w, err := repo.Worktree()
		if err != nil {
			return ClonedRepo{}, fmt.Errorf("failed to get worktree: %w", err)
		}

		pullOpts := &git.PullOptions{}
		token := os.Getenv("GIT_TOKEN")
		if token != "" {
			pullOpts.Auth = &http.BasicAuth{
				Username: "git",
				Password: token,
			}
		}

		err = w.Pull(pullOpts)
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return ClonedRepo{}, fmt.Errorf("failed to pull repository: %w", err)
		}

		return ClonedRepo{
			RepoUrl: repoURL,
			Path:    path,
			repo:    repo,
		}, nil
	} else {
		// Repository doesn't exist, clone it using the existing CloneRepo method
		clonedRepo, err := c.CloneRepo(repoURL)
		if err != nil {
			return ClonedRepo{}, err
		}
		return clonedRepo, nil
	}

	return ClonedRepo{}, nil
}
