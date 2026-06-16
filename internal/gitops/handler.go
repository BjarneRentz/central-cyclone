package gitops

import (
	"central-cyclone/internal/analyzer"
	"central-cyclone/internal/config"
	"central-cyclone/internal/gittool"
	"central-cyclone/internal/upload"
	"context"
	"fmt"
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

func NewCreateSbomChangeHandler(configProvider *config.ConfigProvider, gitTool gittool.Cloner, sbomAnalyzer analyzer.Analyzer, dependencyTrackUploader upload.Uploader) *CreateSbomChangeHandler {
	return &CreateSbomChangeHandler{
		configProvider:          configProvider,
		gitTool:                 gitTool,
		sbomAnalyzer:            sbomAnalyzer,
		dependencyTrackUploader: dependencyTrackUploader,
	}
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
		return fmt.Errorf("get application repo %q: %w", applicationName, err)
	}
	// Clone or update the repo and checkout the specific version
	clonedRepo, err := h.gitTool.CloneOrUpdateRepo(appRepoUrl)
	if err != nil {
		return fmt.Errorf("clone repo %q: %w", appRepoUrl, err)
	}

	err = clonedRepo.CheckoutRevision(version)
	if err != nil {
		return fmt.Errorf("checkout %q: %w", version, err)
	}

	scanTarget, err := h.configProvider.GetScanTargetForApplication(applicationName, environment)
	if err != nil {
		return fmt.Errorf("get scan target %q/%q: %w", applicationName, environment, err)
	}

	sbom, err := h.sbomAnalyzer.AnalyzeProject(clonedRepo, scanTarget)
	if err != nil {
		return fmt.Errorf("analyze %q/%q: %w", applicationName, environment, err)
	}

	err = h.dependencyTrackUploader.UploadSBOM(ctx, sbom)
	if err != nil {
		return fmt.Errorf("upload SBOM %q/%q: %w", applicationName, environment, err)
	}

	return nil
}
