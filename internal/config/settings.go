package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Settings struct {
	Repositories    []Repo                `json:"repositories"`
	DependencyTrack DependencyTrackConfig `json:"dependencyTrack"`
}

type Repo struct {
	Url     string       `json:"url"`
	Targets []RepoTarget `json:"targets"`
}

type RepoTarget struct {
	ProjectId string `json:"projectId"`
	Type      string `json:"type"`
}

type DependencyTrackConfig struct {
	Url string `json:"url"`
}

func LoadFromFile(filePath string) (*Settings, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		return nil, err
	}
	var settings Settings
	err = json.Unmarshal(data, &settings)
	if err != nil {
		fmt.Printf("Error parsing config file: %v\n", err)
		return nil, err
	}
	return &settings, nil
}
