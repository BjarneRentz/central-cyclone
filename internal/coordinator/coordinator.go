package coordinator

import (
	"central-cyclone/internal/analyzer"
	"central-cyclone/internal/config"
	"central-cyclone/internal/upload"
	"central-cyclone/internal/workspace"
	"fmt"
)

func RunForSettings(settings *config.Settings) {

	var workspaceHandler, err = workspace.CreateWorkspace()
	if err != nil {
		fmt.Printf("Error creating workspace: %v\n", err)
		return
	}
	err = workspaceHandler.Clear()
	if err != nil {
		fmt.Printf("Error clearing workspace: %v\n", err)
		return
	}

	if settings != nil && len(settings.Repositories) != 0 {
		uploader := upload.DependencyTrackUploader{ServerURL: settings.DependencyTrack.Url}
		analyzeRepos(settings.Repositories, workspaceHandler, uploader)
	}

}

func analyzeRepos(repoSettings []config.Repo, workspaceHandler workspace.Workspace, uploader upload.Uploader) {
	fmt.Printf("Found %d repositories to analyze 🚀\n", len(repoSettings))

	for _, repo := range repoSettings {
		err := analyzeRepo(&repo, workspaceHandler, uploader)
		if err != nil {
			fmt.Printf("Error analyzing repo %s: %v\n", repo.Url, err)
		}
	}

}

func uploadSbom(uploader upload.Uploader, sbomPath string, projectId string) error {
	err := uploader.UploadSBOM(sbomPath, projectId)
	if err != nil {
		fmt.Printf("Error uploading SBOM: %v\n", err)
		return err
	}
	fmt.Print("⬆️  Uploaded SBOM successfully\n")
	return nil
}

func analyzeRepo(repo *config.Repo, workspaceHandler workspace.Workspace, uploader upload.Uploader) error {
	fmt.Printf("🔎 Analyzing repository: %s\n", repo.Url)

	repoPath, err := workspaceHandler.CloneRepoToWorkspace(repo.Url)
	if err != nil {
		return fmt.Errorf("error cloning repository: %w", err)
	}

	an := analyzer.CdxgenAnalyzer{}

	for _, t := range repo.Targets {
		fmt.Printf("🔬 Analyzing repo for target: %s\n", t.Type)
		sbomPath, err := an.AnalyzeProject(repoPath, t.Type)

		if err != nil {
			return fmt.Errorf("error analyzing project: %v", err)
		}

		err = uploadSbom(uploader, sbomPath, t.ProjectId)
		if err != nil {
			return err
		}

	}
	fmt.Printf("✅ Finished analyzing repo %s\n", repo.Url)
	return nil
}
