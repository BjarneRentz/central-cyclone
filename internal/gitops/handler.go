package gitops

import (
	"central-cyclone/internal/analyzer"
	"central-cyclone/internal/config"
	"central-cyclone/internal/gittool"
	"central-cyclone/internal/upload"
	"context"
	"log/slog"
)

type AppChangedHandler interface {
	HandleAppChange(ctx context.Context, applicationName, environment, version string) error
}

type NoOpsAppChangedHandler struct{}

func (h NoOpsAppChangedHandler) HandleAppChange(ctx context.Context, applicationName, environment, version string) error {
	slog.Debug("Handled App Change", "app", applicationName, "env", environment, "version", version)
	return nil
}

// Creates a new SBOM for the given version
type CreateSbomChangeHandler struct {
	configProvider          *config.ConfigProvider
	gitTool                 gittool.Cloner
	sbomAnalyzer            analyzer.Analyzer
	dependencyTrackUploader upload.Uploader
}

func (h CreateSbomChangeHandler) HandleAppChange(ctx context.Context, applicationName, environment, version string) error {

	appRepoUrl, err := h.configProvider.GetApplicationRepo(applicationName)
	if err != nil {
		slog.Error("Failed to get application repo URL", "application", applicationName, "error", err)
		return err
	}
	// Clone or update the repo and checkout the specific version
	clonedRepo, err := h.gitTool.CloneOrUpdateRepo(appRepoUrl)
	if err != nil {
		slog.Error("Failed to clone or update application repo", "repoUrl", appRepoUrl, "version", version, "error", err)
		return err
	}

	err = clonedRepo.CheckoutTag(version)
	if err != nil {
		return err
	}

	scanTarget, err := h.configProvider.GetScanTargetForApplication(applicationName, environment)
	if err != nil {
		slog.Error("Failed to get scan target for application and environment", "application", applicationName, "environment", environment, "error", err)
		return err
	}

	sbom, err := h.sbomAnalyzer.AnalyzeProject(clonedRepo, scanTarget)
	if err != nil {
		slog.Error("Failed to analyze project", "application", applicationName, "environment", environment, "error", err)
		return err
	}

	err = h.dependencyTrackUploader.UploadSBOM(sbom)
	if err != nil {
		slog.Error("Failed to upload SBOM to DependencyTrack", "application", applicationName, "environment", environment, "error", err)
		return err
	}

	return nil
}
