package gittool

import (
	"fmt"
	"os"
	"os/exec"
)

func CloneRepoToDir(repoURL, targetDir string) error {
	fmt.Printf("🛠️  Cloning repo %s into the workfolder\n", repoURL)
	cmd := exec.Command("git", "clone", "--quiet", repoURL, targetDir)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
