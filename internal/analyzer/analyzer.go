package analyzer

import (
	"central-cyclone/internal/config"
	"central-cyclone/internal/models"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

type Analyzer interface {
	AnalyzeProject(repo models.ClonedRepo, target config.RepoTarget) (models.Sbom, error)
}

type CdxgenAnalyzer struct{}

func (a CdxgenAnalyzer) AnalyzeProject(repo models.ClonedRepo, target config.RepoTarget) (models.Sbom, error) {

	sbomFileName := fmt.Sprintf("sbom_%s.json", target.Type)
	sbomFilePath := filepath.Join(repo.Path, sbomFileName)

	cmd := exec.Command("cdxgen", "--fail-on-error", "-t", target.Type, "-o", sbomFileName)

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
	os.Remove(sbomFilePath)

	if err != nil {
		slog.Error("Failed to read created sbom file", "path", sbomFileName)
		return models.Sbom{}, fmt.Errorf("failed to read sbom file: %v", err)
	}
	return models.Sbom{
		ProjectId:         target.ProjectId,
		ProjectType:       target.Type,
		ProjectFolderName: repo.FolderName,
		Data:              bytes,
	}, nil
}
