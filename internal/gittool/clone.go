package gittool

import (
	"central-cyclone/internal/models"
	"central-cyclone/internal/workspace"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
)

type Cloner interface {
	CloneRepo(repoURL string) (models.ClonedRepo, error)
}

func CreateLocalGitCloner(workspaceHandler workspace.Workspace) Cloner {
	return LocalGitCloner{workspace: workspaceHandler}
}

type LocalGitCloner struct {
	workspace workspace.Workspace
}

func (c LocalGitCloner) CloneRepo(repoURL string) (models.ClonedRepo, error) {
	path, err := c.workspace.CreateRepoFolder(repoURL)
	if err != nil {
		return models.ClonedRepo{}, err
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
		return models.ClonedRepo{}, err
	}

	return models.ClonedRepo{
		RepoUrl: repoURL,
		Path:    path,
	}, nil
}
