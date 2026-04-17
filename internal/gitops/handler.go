package gitops

import (
	"central-cyclone/internal/analyzer"
	"central-cyclone/internal/config"
	"central-cyclone/internal/gittool"
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
	configProvider *config.ConfigProvider
	gitTool        gittool.Cloner
	sbomAnalyzer   analyzer.Analyzer
}

func (h CreateSbomChangeHandler) HandleAppChange(ctx context.Context, applicationName, environment, version string) error {

	// 0. Get AppRepo from applicationName
	appRepoUrl, err := h.configProvider.GetApplicationRepo(applicationName)
	if err != nil {
		slog.Error("Failed to get application repo URL", "application", applicationName, "error", err)
		return err
	}
	//clonedRepo, err := h.gitTool.CloneRepo(appRepoUrl)
	if err != nil {
		slog.Error("Failed to clone application repo", "repoUrl", appRepoUrl, "error", err)
		return err
	}
	//h.sbomAnalyzer.AnalyzeProject(clonedRepo)

	// 3. Updload to corresponding DependencyTrack Project

	return nil
}
