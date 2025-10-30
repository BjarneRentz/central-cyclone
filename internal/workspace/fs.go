package workspace

import (
	"fmt"
	"os"
)

type FSHelper interface {
	CreateFolderIfNotExists(path string) error
	ListFiles(path string) ([]os.DirEntry, error)
	RemoveAll(path string) error
	WriteFile(path string, data []byte) error
}

type LocalFSHelper struct{}

func (h LocalFSHelper) CreateFolderIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create folder '%s': %w", path, err)
		}
	}
	return nil
}

func (h LocalFSHelper) ListFiles(path string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory '%s': %w", path, err)
	}
	return entries, nil
}

func (h LocalFSHelper) RemoveAll(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove '%s': %w", path, err)
	}
	return nil
}

func (h LocalFSHelper) WriteFile(path string, data []byte) error {
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file '%s': %w", path, err)
	}
	return nil
}
