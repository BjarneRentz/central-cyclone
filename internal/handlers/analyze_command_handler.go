package handlers

import (
	"central-cyclone/internal/analyzer"
	"central-cyclone/internal/config"
	"central-cyclone/internal/gittool"
	"central-cyclone/internal/models"
	"central-cyclone/internal/upload"
	"central-cyclone/internal/workspace"
	"fmt"
	"log/slog"
)

func AnalyzeAndSave(settings *config.Settings, gitTool gittool.Cloner, workspaceHandler workspace.Workspace) {
	if settings != nil && len(settings.Repositories) != 0 {
		analyzeRepos(settings.Repositories, gitTool, workspaceHandler, nil)
	}
}

func AnalyzeAndUpload(settings *config.Settings, gitTool gittool.Cloner, workspaceHandler workspace.Workspace, uploader upload.Uploader) {
	if settings != nil && len(settings.Repositories) != 0 {
		analyzeRepos(settings.Repositories, gitTool, workspaceHandler, uploader)
	}
}

func analyzeRepos(repoSettings []config.Repo, gitTool gittool.Cloner, workspaceHandler workspace.Workspace, uploader upload.Uploader) {
	slog.Info("Found repositories to analyze", "count", len(repoSettings))

	for _, repo := range repoSettings {
		err := analyzeRepo(&repo, gitTool, workspaceHandler, uploader)
		if err != nil {
			slog.Error("Could not analyze repo", "repo", repo.Url, "error", err)
		}
	}

}

func uploadSbom(uploader upload.Uploader, sbom models.Sbom) error {
	err := uploader.UploadSBOM(sbom)
	if err != nil {
		slog.Error("Could not upload SBOM", "error", err)
		return err
	}
	return nil
}

func analyzeRepo(repo *config.Repo, gitTool gittool.Cloner, workspaceHandler workspace.Workspace, uploader upload.Uploader) error {
	slog.Info("🔎 Analyzing repository", "repo", repo.Url)

	analyzer := analyzer.CdxgenAnalyzer{}

	clonedRepo, err := gitTool.CloneRepo(repo.Url)
	if err != nil {
		slog.Error("Could not clone repository", "repo", repo.Url, "error", err)
		return fmt.Errorf("error cloning repository: %w", err)
	}

	for _, t := range repo.Targets {
		slog.Info("🔬 Analyzing repo", "repo", repo.Url, "target", t.Type)
		sbom, err := analyzer.AnalyzeProject(clonedRepo, t)
		if err != nil {
			return fmt.Errorf("error analyzing project: %v", err)
		}

		if uploader != nil {
			_ = uploadSbom(uploader, sbom)
		} else {
			err := workspaceHandler.SaveSbom(sbom)
			if err != nil {
				slog.Error("Could not save sbom", "error", err)
				return fmt.Errorf("error saving sbom: %v", err)
			}
		}

	}
	slog.Info("✅ Finished analyzing repo", "repo", repo.Url)
	return nil
}
