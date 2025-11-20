package workspace

import (
	"central-cyclone/internal/sbom"
	"fmt"
	"path/filepath"
	"strings"
)

type SBOMNamer interface {
	GenerateSBOMPath(sbomsDir string, sbom sbom.Sbom) string
	ParseFilename(filename string) (repoFolderName string, projectType string, err error)
}

type DefaultSBOMNamer struct{}

func (n DefaultSBOMNamer) GenerateSBOMPath(sbomsDir string, sbom sbom.Sbom) string {
	sbomFileName := fmt.Sprintf("%s_sbom_%s.json", sbom.ProjectFolderName, sbom.ProjectType)
	return filepath.Join(sbomsDir, sbomFileName)
}

func (n DefaultSBOMNamer) ParseFilename(fileName string) (repoFolderName string, projectType string, err error) {
	fileNameWithoutExt := strings.TrimSuffix(fileName, ".json")

	parts := strings.Split(fileNameWithoutExt, "_sbom_")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid filename format: %s expected format: org_repo_sbom_type.json", fileName)
	}

	folderName := parts[0]
	projectType = parts[1]

	return folderName, projectType, nil
}
