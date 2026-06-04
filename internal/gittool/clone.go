package gittool

import (
	"central-cyclone/internal/workspace"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/transport"
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
		slog.Info("📁 Repository already exists, fetching updates", "repo", repoURL)
		repo, err := git.PlainOpen(path)
		if err != nil {
			return ClonedRepo{}, fmt.Errorf("failed to open existing repository: %w", err)
		}

		// Setup authentication
		var auth transport.AuthMethod
		token := os.Getenv("GIT_TOKEN")
		if token != "" {
			auth = &http.BasicAuth{
				Username: "git",
				Password: token,
			}
		}

		// Fetch updates from the remote (updates all tags and remote branches)
		err = repo.Fetch(&git.FetchOptions{
			Auth:  auth,
			Force: true,        // Ensures tags are updated/overwritten if changed on remote
			Tags:  git.AllTags, // Explicitly pull down all tags
		})

		// git.NoErrAlreadyUpToDate means there was nothing new to download, which is perfectly fine!
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return ClonedRepo{}, fmt.Errorf("failed to fetch repository updates: %w", err)
		}

		return ClonedRepo{
			RepoUrl: repoURL,
			Path:    path,
			repo:    repo,
		}, nil
	} else {
		// Repository doesn't exist, clone it
		clonedRepo, err := c.CloneRepo(repoURL)
		if err != nil {
			return ClonedRepo{}, err
		}
		return clonedRepo, nil
	}
}
