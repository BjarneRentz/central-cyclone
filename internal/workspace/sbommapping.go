package workspace

import (
	"central-cyclone/internal/config"
	"central-cyclone/internal/sbom"
	"fmt"
	"path/filepath"
)

type SBOMFileName struct {
	FolderName  string // e.g., "org_repo"
	ProjectType string // e.g., "go", "npm"
}

type SBOMNamer interface {
	GenerateSBOMPath(sbomsDir string, sbom sbom.Sbom) string
	MapSBOMToProject(settings *config.Settings, folderName, projectType string) (projectId string, found bool)
}

type DefaultSBOMNamer struct{}

func (n DefaultSBOMNamer) GenerateSBOMPath(sbomsDir string, sbom sbom.Sbom) string {
	sbomFileName := fmt.Sprintf("%s_sbom_%s.json", sbom.ProjectFolderName, sbom.ProjectType)
	return filepath.Join(sbomsDir, sbomFileName)
}

func (n DefaultSBOMNamer) MapSBOMToProject(settings *config.Settings, folderName, projectType string) (string, bool) {
	mapper := DefaultRepoMapper{}
	for _, repo := range settings.Repositories {
		repoFolder, err := mapper.GetFolderName(repo.Url)
		if err != nil {
			continue
		}
		if repoFolder != folderName {
			continue
		}

		// Find target with matching type
		for _, target := range repo.Targets {
			if target.Type == projectType {
				return target.ProjectId, true
			}
		}
		break // found repo but no matching target
	}
	return "", false
}
