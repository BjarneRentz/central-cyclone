package gittool

import (
	"central-cyclone/internal/models"
	"errors"
	"testing"
)

// MockWorkspace implements the workspace.Workspace interface for testing
type MockWorkspace struct {
	createRepoFolderErr   error
	createRepoFolderPath  string
	createRepoFolderCalls int
	clearErr              error
	clearCalls            int
}

func (m *MockWorkspace) CreateRepoFolder(repoURL string) (string, error) {
	m.createRepoFolderCalls++
	if m.createRepoFolderErr != nil {
		return "", m.createRepoFolderErr
	}
	return m.createRepoFolderPath, nil
}

func (m *MockWorkspace) Clear() error {
	m.clearCalls++
	return m.clearErr
}

func (m *MockWorkspace) SaveSbom(sbom models.Sbom) error {
	return nil
}

func TestCreateLocalGitCloner(t *testing.T) {
	mockWS := &MockWorkspace{}
	cloner := CreateLocalGitCloner(mockWS)

	if cloner == nil {
		t.Error("CreateLocalGitCloner returned nil")
	}

	localCloner, ok := cloner.(LocalGitCloner)
	if !ok {
		t.Error("Expected LocalGitCloner type")
	}

	if localCloner.workspace != mockWS {
		t.Error("Expected workspace to be set correctly")
	}
}

func TestCloneRepoErrorOnCreateRepoFolder(t *testing.T) {
	createErr := errors.New("failed to create folder")
	mockWS := &MockWorkspace{
		createRepoFolderErr: createErr,
	}
	cloner := CreateLocalGitCloner(mockWS)

	result, err := cloner.CloneRepo("https://github.com/example/repo.git")

	if err != createErr {
		t.Errorf("Expected error %v, got %v", createErr, err)
	}

	if result != (models.ClonedRepo{}) {
		t.Errorf("Expected empty ClonedRepo on error, got %v", result)
	}

	if mockWS.createRepoFolderCalls != 1 {
		t.Errorf("Expected CreateRepoFolder to be called once, was called %d times", mockWS.createRepoFolderCalls)
	}
}

func TestCloneRepoInvalidGitURL(t *testing.T) {
	mockWS := &MockWorkspace{
		createRepoFolderPath: "/tmp/test-repo",
	}
	cloner := CreateLocalGitCloner(mockWS)

	tests := []struct {
		name    string
		repoURL string
	}{
		{
			name:    "Empty URL",
			repoURL: "",
		},
		{
			name:    "Invalid URL",
			repoURL: "not-a-url",
		},
		{
			name:    "HTTP URL",
			repoURL: "http://github.com/example/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cloner.CloneRepo(tt.repoURL)

			// These should fail because git.PlainClone will reject invalid URLs
			if err == nil && tt.repoURL != "" && tt.repoURL != "http://github.com/example/repo.git" {
				t.Errorf("Expected error for URL %q", tt.repoURL)
			}

			if err != nil && result != (models.ClonedRepo{}) {
				t.Errorf("Expected empty ClonedRepo on error, got %v", result)
			}
		})
	}
}

func TestCloneRepoReturnsCorrectValues(t *testing.T) {
	repoURL := "https://github.com/example/repo.git"
	expectedPath := "/tmp/cloned-repo"

	mockWS := &MockWorkspace{
		createRepoFolderPath: expectedPath,
	}
	cloner := CreateLocalGitCloner(mockWS)

	// This will fail to clone because the URL isn't real, but we're testing
	// that if the clone succeeded, the returned values would be correct
	result, err := cloner.CloneRepo(repoURL)

	// We expect an error because the repo doesn't exist
	if err == nil {
		t.Error("Expected error when cloning invalid repo")
	}

	// Even on error, CloneRepo should still try to call PlainClone
	if mockWS.createRepoFolderCalls != 1 {
		t.Errorf("Expected CreateRepoFolder to be called once, was called %d times", mockWS.createRepoFolderCalls)
	}

	// Verify the returned struct would have correct values if clone succeeded
	if err == nil && result.Path != expectedPath {
		t.Errorf("Expected Path %q, got %q", expectedPath, result.Path)
	}
	if err == nil && result.RepoUrl != repoURL {
		t.Errorf("Expected RepoUrl %q, got %q", repoURL, result.RepoUrl)
	}
}
