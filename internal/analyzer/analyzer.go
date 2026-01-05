package analyzer

import (
	"central-cyclone/internal/config"
	"central-cyclone/internal/sbom"
	"central-cyclone/internal/workspace"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

type Analyzer interface {
	AnalyzeProject(repo workspace.ClonedRepo, target config.RepoTarget) (sbom.Sbom, error)
}

type CdxgenAnalyzer struct{}

func (a CdxgenAnalyzer) AnalyzeProject(repo workspace.ClonedRepo, target config.RepoTarget) (sbom.Sbom, error) {

	fileName := fmt.Sprintf("sbom_%s.json", target.Type)

	scanPath := repo.Path

	if target.Directory != nil {
		scanPath = filepath.Join(repo.Path, *target.Directory)
	}

	cmd := exec.Command("cdxgen", "--fail-on-error", "-t", target.Type, "-o", fileName, scanPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("Creating sbom with cdxgen failed: ", "output", string(output), "error", err)
		return sbom.Sbom{}, fmt.Errorf("cdxgen failed: %v\nOutput: %s", err, string(output))
	}

	bytes, err := os.ReadFile(fileName)
	os.Remove(fileName)

	if err != nil {
		slog.Error("Failed to read created sbom file", "path", fileName)
		return sbom.Sbom{}, fmt.Errorf("failed to read sbom file: %v", err)
	}
	return sbom.Sbom{
		ProjectId:         target.ProjectId,
		ProjectType:       target.Type,
		ProjectFolderName: repo.FolderName,
		Data:              bytes,
	}, nil
}
