package gittool

import (
	"fmt"
	"os"
	"os/exec"
)

type Cloner interface {
	CloneRepoToDir(repoURL, targetDir string) error
}

type LocalGitCloner struct {
}

func (c LocalGitCloner) CloneRepoToDir(repoURL, targetDir string) error {
	fmt.Printf("🛠️  Cloning repo %s into the workfolder\n", repoURL)
	cmd := exec.Command("git", "clone", "--quiet", repoURL, targetDir)
	cmd.Stderr = os.Stderr
	return cmd.Run()

}
