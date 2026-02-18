package config

import (
	"encoding/json"
	"log/slog"
	"os"
)

type Settings struct {
	Repositories     []Repo                `json:"repositories"`
	DependencyTrack  DependencyTrackConfig `json:"dependencyTrack"`
	Applications     []Application         `json:"applications"`
	GitOpsRepos      []GitOpsRepo          `json:"gitOpsRepos"`
	ApllicationRepos []ApllicationRepo     `json:"applicationRepos"`
}

type Repo struct {
	Url     string       `json:"url"`
	Targets []RepoTarget `json:"targets"`
}

type RepoTarget struct {
	ProjectId string  `json:"projectId"`
	Type      string  `json:"type"`
	Directory *string `json:"directory"`
}

type DependencyTrackConfig struct {
	Url string `json:"url"`
}

type Application struct {
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	Projects []Project `json:"projects"`
}

type Project struct {
	Name      string  `json:"name"`
	Version   string  `json:"version"`
	IsLatest  bool    `json:"isLatest"`
	ProjectId *string `json:"projectId"`
}

type GitOpsRepo struct {
	Url                string              `json:"url"`
	GitOpsApplications []GitOpsApplication `json:"gitOpsApplications"`
}

type GitOpsApplication struct {
	ApplicationName    string              `json:"applicationName"`
	VersionIdentifiers []VersionIdentifier `json:"versionIdentifiers"`
}

type VersionIdentifier struct {
	Environment string `json:"environment"`
	Filepath    string `json:"filepath"`
	YamlPath    string `json:"yamlPath"`
}

type ApllicationRepo struct {
	Applications []string `json:"applications"`
	RepoUrl      string   `json:"repoUrl"`
}

func LoadFromFile(filePath string) (*Settings, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var settings Settings
	err = json.Unmarshal(data, &settings)
	if err != nil {
		slog.Error("Error parsing config file:", "error", err)
		return nil, err
	}
	return &settings, nil
}
