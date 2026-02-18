package workspace

import (
	"central-cyclone/internal/sbom"
	"fmt"
	"path/filepath"
)

type SBOMNamer interface {
	GenerateSBOMPath(sbomsDir string, sbom sbom.Sbom) string
}

type DefaultSBOMNamer struct{}

func (n DefaultSBOMNamer) GenerateSBOMPath(sbomsDir string, sbom sbom.Sbom) string {
	sbomFileName := fmt.Sprintf("sbom_%s.json", sbom.ProjectId)
	return filepath.Join(sbomsDir, sbomFileName)
}
