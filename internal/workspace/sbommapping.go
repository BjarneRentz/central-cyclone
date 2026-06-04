package workspace

import (
	"central-cyclone/internal/models"
	"fmt"
	"path/filepath"
)

type SBOMNamer interface {
	GenerateSBOMPath(sbomsDir string, sbom models.Sbom) string
}

type DefaultSBOMNamer struct{}

func (n DefaultSBOMNamer) GenerateSBOMPath(sbomsDir string, sbom models.Sbom) string {
	sbomFileName := fmt.Sprintf("sbom_%s.json", sbom.ProjectId)
	return filepath.Join(sbomsDir, sbomFileName)
}
