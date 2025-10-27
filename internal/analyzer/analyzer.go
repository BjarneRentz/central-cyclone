package analyzer

import (
	"fmt"
	"os/exec"
)

type Analyzer interface {
	AnalyzeProject(projectPath string, projectType string, sbomPath string) error
}

type CdxgenAnalyzer struct{}

func (a CdxgenAnalyzer) AnalyzeProject(projectPath string, projectType string, sbomPath string) error {
	cmd := exec.Command("cdxgen", "--fail-on-error", "-t", projectType, "-o", sbomPath, projectPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("cdxgen failed with %s \n", string(output))
		return fmt.Errorf("cdxgen failed: %v\nOutput: %s", err, string(output))
	}
	return nil
}
