package gittool

import (
	"os"
	"os/exec"
)

// CloneRepoToDir clones a git repo from the given URL into the specified directory.
func CloneRepoToDir(repoURL, targetDir string) error {
	cmd := exec.Command("git", "clone", repoURL, targetDir)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
