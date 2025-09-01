package config

import (
	"os"
	"testing"
)

func TestLoadFromFile_InvalidFile(t *testing.T) {
	// Arrange & Act
	_, err := LoadFromFile("/invalid/path/config.json")

	// Assert
	if err == nil {
		t.Error("expected error for invalid file path, got nil")
	}
}

func TestLoadFromFile_ValidFile(t *testing.T) {
	// Arrange
	fileContent := `{"repositories":[], "dependencyTrack":{"url":"http://localhost"}}`
	tmpFile, err := os.CreateTemp("", "config.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.Write([]byte(fileContent))
	if err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()
	// Act
	settings, err := LoadFromFile(tmpFile.Name())

	// Assert
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if settings.DependencyTrack.Url != "http://localhost" {
		t.Errorf("unexpected DependencyTrack.Url: %s", settings.DependencyTrack.Url)
	}
}
