package gittool

import (
	"log/slog"
	"os"
	"os/exec"
)

type Cloner interface {
	CloneRepoToDir(repoURL, targetDir string) error
}

type LocalGitCloner struct {
}

func (c LocalGitCloner) CloneRepoToDir(repoURL, targetDir string) error {
	slog.Info("🛠️  Cloning repo into the workfolder", "repo", repoURL)
	repoURL = adaptUrlIfTokenIsProvided(repoURL)
	cmd := exec.Command("git", "clone", "--quiet", "--depth", "1", repoURL, targetDir)
	cmd.Stderr = os.Stderr
	return cmd.Run()

}

func adaptUrlIfTokenIsProvided(repoURL string) string {
	token := os.Getenv("GIT_TOKEN")
	if token != "" && len(repoURL) > 8 && repoURL[:8] == "https://" {
		repoURL = "https://" + token + "@" + repoURL[8:]
	}
	return repoURL
}
