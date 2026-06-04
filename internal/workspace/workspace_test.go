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
	w := localWorkspace{
		path: "./nonexistentdir",
		fs:   LocalFSHelper{},
	}
	err := w.Clear()
	if err != nil {
		t.Error("expected no error for non-existent directory")
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

			// Create a workspace with the mock cloner
			w := localWorkspace{
				path:       tempDir,
				reposPath:  reposPath,
				fs:         LocalFSHelper{},
				repoMapper: DefaultRepoMapper{},
			}

			// Call CloneRepoToWorkspace
			repoPath, err := w.CreateRepoFolder(tt.repoURL)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify the returned path matches expected
			if repoPath != tt.wantPath {
				t.Errorf("CloneRepoToWorkspace() returned path = %v, want %v", repoPath, tt.wantPath)
			}

			// Verify the directory was created
			if _, err := os.Stat(tt.wantPath); os.IsNotExist(err) {
				t.Errorf("expected directory %v to be created", tt.wantPath)
			}
		})
	}
}
