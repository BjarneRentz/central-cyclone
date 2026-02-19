package gittool

import (
	"central-cyclone/internal/models"
	"central-cyclone/internal/workspace"
	"log/slog"
	"os"
	"os/exec"
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
	repoURL = adaptUrlIfTokenIsProvided(repoURL)
	cmd := exec.Command("git", "clone", "--quiet", "--depth", "1", repoURL, path)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return models.ClonedRepo{}, err
	}
	return models.ClonedRepo{
		RepoUrl: repoURL,
		Path:    path,
	}, nil
}

func adaptUrlIfTokenIsProvided(repoURL string) string {
	token := os.Getenv("GIT_TOKEN")
	if token != "" && len(repoURL) > 8 && repoURL[:8] == "https://" {
		repoURL = "https://" + token + "@" + repoURL[8:]
	}
	return repoURL
}
