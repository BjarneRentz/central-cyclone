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

	for _, repo := range *repoSettings {
		analyzeRepo(&repo, workspaceHandler)
	}

}

func analyzeRepo(repo *config.Repo, workspaceHandler workspace.Workspace) {
	fmt.Printf("🔎 Analyzing repository: %s\n", repo.Url)

	repoPath, err := workspaceHandler.CloneRepoToWorkspace(repo.Url)
	if err != nil {
		fmt.Printf("Error cloning repository: %v\n", err)
		return
	}

	an := analyzer.CdxgenAnalyzer{}
	fmt.Printf("Analyzing %s...\n", repoPath)

	for _, t := range repo.Targets {
		fmt.Printf("🔬 Analyzing repo for target: %s\n", t.Type)
		_, err = an.AnalyzeProject(repoPath, t.Type)

		if err != nil {
			fmt.Printf("Error analyzing project: %v\n", err)
			return
		}
	}
	fmt.Printf("✅ Finished analyzing repo %s\n", repo.Url)
}
