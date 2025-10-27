package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

type mockGitCloner struct {
	lastRepoURL   string
	lastTargetDir string
}

func (m *mockGitCloner) CloneRepoToDir(repoURL, targetDir string) error {
	m.lastRepoURL = repoURL
	m.lastTargetDir = targetDir

	return nil
}

func TestWorkspaceHandler_Clear_NonExistent_DoesNotThrow(t *testing.T) {
	w := localWorkspace{path: "./nonexistentdir"}
	err := w.Clear()
	if err != nil {
		t.Error("expected no error for non-existent directory")
	}
}

func TestWorkspaceHandler_Clear_Empty(t *testing.T) {
	dir := t.TempDir()
	w := localWorkspace{path: dir}
	if err := w.Clear(); err != nil {
		t.Errorf("unexpected error clearing empty dir: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Errorf("failed to read dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty dir, found %d entries", len(entries))
	}
}

func TestCloneRepoToWorkspace(t *testing.T) {
	tempDir := t.TempDir()
	reposPath := filepath.Join(tempDir, "repos")

	tests := []struct {
		name     string
		repoURL  string
		wantPath string
	}{
		{
			name:     "GitHub repository",
			repoURL:  "https://github.com/org/repo",
			wantPath: filepath.Join(reposPath, "org_repo"),
		},
		{
			name:     "Azure DevOps repository",
			repoURL:  "https://dev.azure.com/my-org/my-project/_git/my-repo",
			wantPath: filepath.Join(reposPath, "my-org_my-project_my-repo"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock git cloner
			mockCloner := &mockGitCloner{}

			// Create a workspace with the mock cloner
			w := localWorkspace{
				path:      tempDir,
				reposPath: reposPath,
				gitCloner: mockCloner,
			}

			// Call CloneRepoToWorkspace
			gotPath, err := w.CloneRepoToWorkspace(tt.repoURL)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify the returned path matches expected
			if gotPath != tt.wantPath {
				t.Errorf("CloneRepoToWorkspace() returned path = %v, want %v", gotPath, tt.wantPath)
			}

			// Verify that CloneRepoToDir was called with correct arguments
			if mockCloner.lastRepoURL != tt.repoURL {
				t.Errorf("CloneRepoToDir() was called with repoURL = %v, want %v", mockCloner.lastRepoURL, tt.repoURL)
			}
			if mockCloner.lastTargetDir != tt.wantPath {
				t.Errorf("CloneRepoToDir() was called with targetDir = %v, want %v", mockCloner.lastTargetDir, tt.wantPath)
			}

			// Verify the directory was created
			if _, err := os.Stat(tt.wantPath); os.IsNotExist(err) {
				t.Errorf("expected directory %v to be created", tt.wantPath)
			}
		})
	}
}

func TestGetFolderNameForRepoUrl(t *testing.T) {
	var tests = []struct {
		repoUrl            string
		expectedFolderName string
	}{
		{"https://github.com/org/repo.git", "org_repo"},
		{"https://dev.azure.com/my-org/my-project/_git/my-repo", "my-org_my-project_my-repo"},
	}

	for _, tt := range tests {
		t.Run(tt.repoUrl, func(t *testing.T) {
			ans, _ := getFolderNameForRepoUrl(tt.repoUrl)
			if ans != tt.expectedFolderName {
				t.Errorf("got %s, want %s", ans, tt.expectedFolderName)
			}
		})
	}
}
