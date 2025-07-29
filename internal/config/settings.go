package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Settings struct {
	Repositories []Repo `json:"repositories"`
}

type Repo struct {
	Url       string `json:"url"`
	ProjectId string `json:"projectId"`
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
