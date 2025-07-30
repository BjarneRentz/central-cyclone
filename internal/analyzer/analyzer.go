package analyzer

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

type Analyzer interface {
	AnalyzeProject(projectPath string, projectType string) (string, error)
}

type CdxgenAnalyzer struct{}

func (a CdxgenAnalyzer) AnalyzeProject(projectPath string, projectType string) (string, error) {
	sbomPath := filepath.Join(projectPath, "sbom.json")
	cmd := exec.Command("cdxgen", "-t", projectType, "-o", sbomPath, projectPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cdxgen failed: %v\nOutput: %s", err, string(output))
	}
	return sbomPath, nil
}
