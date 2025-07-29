package analyzer

import (
	"central-cyclone/internal/config"
	"central-cyclone/internal/gittool"
	"central-cyclone/internal/workspace"
	"fmt"
)

func Analyze(Settings *config.Settings) {

	var workspaceHandler, err = workspace.CreateWorkspace()
	if err != nil {
		fmt.Printf("Error creating workspace: %v\n", err)
		return
	}

	if Settings != nil && len(Settings.Repositories) != 0 {
		analyzeRepos(&Settings.Repositories, workspaceHandler)
	}

	for _, repo := range Settings.Repositories {
		fmt.Printf("Repository URL: %s\n", repo.Url)
	}
}

func analyzeRepos(repoSettings *[]config.Repo, workspaceHandler workspace.WorkspaceHandler) {
	fmt.Printf("Found %d repositories to analyze 🚀\n", len(*repoSettings))

	// Create Workdir if required or clean it up => move it into a own function and not in the git tool
	// clone repos
	// create sboms
	gittool.CloneRepo((*repoSettings)[0].Url)

	workspaceHandler.Clear()

}
