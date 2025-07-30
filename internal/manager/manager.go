package manager

import (
	"central-cyclone/internal/analyzer"
	"central-cyclone/internal/config"
	"central-cyclone/internal/workspace"
	"fmt"
)

func RunForSettings(Settings *config.Settings) {

	var workspaceHandler, err = workspace.CreateWorkspace()
	if err != nil {
		fmt.Printf("Error creating workspace: %v\n", err)
		return
	}
	workspaceHandler.Clear()

	if Settings != nil && len(Settings.Repositories) != 0 {
		analyzeRepos(&Settings.Repositories, workspaceHandler)
	}

}

func analyzeRepos(repoSettings *[]config.Repo, workspaceHandler workspace.Workspace) {
	fmt.Printf("Found %d repositories to analyze 🚀\n", len(*repoSettings))

	// clone repos
	// create sboms
	repoPath, err := workspaceHandler.CloneRepoToWorkspace((*repoSettings)[0].Url)
	if err != nil {
		fmt.Printf("Error cloning repository: %v\n", err)
		return
	}

	fmt.Printf("Cloned repository to: %s\n", repoPath)

	an := analyzer.CdxgenAnalyzer{}
	fmt.Printf("Analyzing %s...\n", repoPath)
	_, err = an.AnalyzeProject(repoPath, "node")
	if err != nil {
		fmt.Printf("Error analyzing project: %v\n", err)
	}
	fmt.Printf("✅ Finished analyzing repo %s\n", (*repoSettings)[0].Url)

}
