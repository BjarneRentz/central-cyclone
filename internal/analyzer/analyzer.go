package analyzer

import (
	"central-cyclone/internal/gittool"
	"central-cyclone/internal/models"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

type Analyzer interface {
	AnalyzeProject(repo gittool.ClonedRepo, target *ScanTarget) (models.Sbom, error)
}

type ScanTarget struct {
	ProjectId   string
	ProjectType string
	Directory   *string
}

type CdxgenAnalyzer struct{}

func (a CdxgenAnalyzer) AnalyzeProject(repo gittool.ClonedRepo, target *ScanTarget) (models.Sbom, error) {

	sbomFileName := fmt.Sprintf("sbom_%s.json", target.ProjectType)
	sbomFilePath := filepath.Join(repo.Path, sbomFileName)

	cmd := exec.Command("cdxgen", "--fail-on-error", "-t", target.ProjectType, "-o", sbomFileName)

	if target.Directory != nil {
		cmd.Args = append(cmd.Args, *target.Directory)
	}

	cmd.Dir = repo.Path
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("Creating sbom with cdxgen failed: ", "output", string(output), "error", err)
		return models.Sbom{}, fmt.Errorf("cdxgen failed: %v\nOutput: %s", err, string(output))
	}

	bytes, err := os.ReadFile(sbomFilePath)
	sbomstring := string(bytes)
	os.Remove(sbomFilePath)

	if err != nil {
		slog.Error("Failed to read created sbom file", "path", sbomFileName)
		return models.Sbom{}, fmt.Errorf("failed to read sbom file: %v", err)
	}
	return models.Sbom{
		ProjectId:   target.ProjectId,
		ProjectType: target.ProjectType,
		Data:        sbomstring,
	}, nil
}
