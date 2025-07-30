package analyzer

import (
	"central-cyclone/internal/config"
	"central-cyclone/internal/workspace"
	"fmt"
)

func Analyze(Settings *config.Settings) {

	var workspaceHandler, err = workspace.CreateWorkspace()
	if err != nil {
		fmt.Printf("Error creating workspace: %v\n", err)
		return
	}
	workspaceHandler.Clear()

	if Settings != nil && len(Settings.Repositories) != 0 {
		analyzeRepos(&Settings.Repositories, workspaceHandler)
	}

	for _, repo := range Settings.Repositories {
		fmt.Printf("Repository URL: %s\n", repo.Url)
	}
}

func analyzeRepos(repoSettings *[]config.Repo, workspaceHandler workspace.Workspace) {
	fmt.Printf("Found %d repositories to analyze 🚀\n", len(*repoSettings))

	// clone repos
	// create sboms
	name, err := workspaceHandler.CloneRepoToWorkspace((*repoSettings)[0].Url)
	if err != nil {
		fmt.Printf("Error cloning repository: %v\n", err)
		return
	}

	fmt.Printf("Cloned repository to: %s\n", name)

}
