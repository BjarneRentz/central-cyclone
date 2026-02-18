package workspace

import (
	"central-cyclone/internal/config"
	"central-cyclone/internal/sbom"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
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
// Or have an own file format / json format that holds all required informations. This would allow us to just read and upload file after file
// without requiring the config.
func (w LocalReadonlySbomWorkspace) ReadSboms(repos []config.Repo) ([]sbom.Sbom, error) {

	filePaths, err := w.fs.ListFiles(w.path)
	if err != nil {
		return nil, err
	}

	var sboms []sbom.Sbom

	for _, filePath := range filePaths {
		if !strings.HasSuffix(filePath, ".json") {
			slog.Info("Skipping non-JSON file", "file", filePath)
			continue
		}

		data, err := os.ReadFile(filePath)

		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		sbom := sbom.Sbom{}

		json.Unmarshal(data, &sbom)

		sboms = append(sboms, sbom)
	}

	return sboms, nil
}
