package workspace

import (
	"central-cyclone/internal/config"
	"central-cyclone/internal/sbom"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ReadonlySbomWorkspace interface {
	ReadSboms(repos []config.Repo) ([]sbom.Sbom, error)
}

type LocalReadonlySbomWorkspace struct {
	path       string
	fs         FSHelper
	sbomNamer  SBOMNamer
	repoMapper RepoURLMapper
}

func CreateLocalReadonlySbomWorkspace(path string, sbomNamer SBOMNamer, repoMapper RepoURLMapper) ReadonlySbomWorkspace {
	return LocalReadonlySbomWorkspace{path: path, fs: LocalFSHelper{}, sbomNamer: sbomNamer, repoMapper: repoMapper}
}

// Note: The current behavior is not the best in terms of performance, as it reads all SBOMs at once.
// In the future, we could inject an uploader, such that the garbarge collector can come in earlier.
func (w LocalReadonlySbomWorkspace) ReadSboms(repos []config.Repo) ([]sbom.Sbom, error) {
	repoMap := make(map[string]config.Repo) // Map folder name -> Repo

	for _, repo := range repos {
		folderName, err := w.repoMapper.GetFolderName(repo.Url)
		if err != nil {
			return nil, fmt.Errorf("failed to map repo URL %s: %w", repo.Url, err)
		}
		repoMap[folderName] = repo
	}

	filePaths, err := w.fs.ListFiles(w.path)
	if err != nil {
		return nil, err
	}

	var sboms []sbom.Sbom

	for _, filePath := range filePaths {
		if !strings.HasSuffix(filePath, ".json") {
			fmt.Printf("Skipping non-JSON file: %s\n", filePath)
			continue
		}

		fileName := filepath.Base(filePath)

		repoFolderName, projectType, err := w.sbomNamer.ParseFilename(fileName)
		if err != nil {
			fmt.Printf("Warning: %v\n", err)
			continue
		}

		repo, exists := repoMap[repoFolderName]
		if !exists {
			fmt.Printf("Warning: No repo found for folder %s (file: %s)\n", repoFolderName, filePath)
			continue
		}

		var projectId string
		for _, target := range repo.Targets {
			if target.Type == projectType {
				projectId = target.ProjectId
				break
			}
		}

		if projectId == "" {
			fmt.Printf("Warning: No target found for type %s in repo %s (file: %s)\n", projectType, repoFolderName, filePath)
			continue
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		sbom := sbom.Sbom{
			ProjectId:         projectId,
			ProjectType:       projectType,
			ProjectFolderName: repoFolderName,
			Data:              data,
		}

		sboms = append(sboms, sbom)
	}

	return sboms, nil
}
