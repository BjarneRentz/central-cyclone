package handlers

import (
	"central-cyclone/internal/analyzer"
	"central-cyclone/internal/config"
	"central-cyclone/internal/sbom"
	"central-cyclone/internal/upload"
	"central-cyclone/internal/workspace"
	"fmt"
)

func AnalyzeAndSave(settings *config.Settings, workspaceHandler workspace.Workspace) {
	if settings != nil && len(settings.Repositories) != 0 {
		analyzeRepos(settings.Repositories, workspaceHandler, nil)
	}
}

func AnalyzeAndUpload(settings *config.Settings, workspaceHandler workspace.Workspace, uploader upload.Uploader) {
	if settings != nil && len(settings.Repositories) != 0 {
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

func uploadSbom(uploader upload.Uploader, sbom sbom.Sbom) error {
	err := uploader.UploadSBOM(sbom)
	if err != nil {
		fmt.Printf("Error uploading SBOM: %v\n", err)
		return err
	}
	return nil
}

func analyzeRepo(repo *config.Repo, workspaceHandler workspace.Workspace, uploader upload.Uploader) error {
	fmt.Printf("🔎 Analyzing repository: %s\n", repo.Url)

	analyzer := analyzer.CdxgenAnalyzer{}

	clonedRepo, err := workspaceHandler.CloneRepoToWorkspace(repo.Url)
	if err != nil {
		return fmt.Errorf("error cloning repository: %w", err)
	}

	for _, t := range repo.Targets {
		fmt.Printf("🔬 Analyzing repo for target: %s\n", t.Type)
		sbom, err := analyzer.AnalyzeProject(clonedRepo, t)
		if err != nil {
			return fmt.Errorf("error analyzing project: %v", err)
		}

		if uploader != nil {
			_ = uploadSbom(uploader, sbom)
		} else {
			err := workspaceHandler.SaveSbom(sbom)
			if err != nil {
				fmt.Printf("Could not save sbom: %s \n", err)
				return fmt.Errorf("error saving sbom: %v", err)
			}
		}

	}
	fmt.Printf("✅ Finished analyzing repo %s\n", repo.Url)
	return nil
}
