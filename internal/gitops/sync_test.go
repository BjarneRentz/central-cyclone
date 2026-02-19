package gitops

import (
	"central-cyclone/internal/config"
	"central-cyclone/internal/gittool"
	"central-cyclone/internal/models"
	"errors"
	"testing"
)

// MockCloner implements gittool.Cloner for testing
type MockCloner struct {
	cloneRepoCallCount int
	cloneRepoErr       error
	cloneRepoResult    gittool.ClonedRepo
}

func (m *MockCloner) CloneRepo(repoURL string) (gittool.ClonedRepo, error) {
	m.cloneRepoCallCount++
	if m.cloneRepoErr != nil {
		return gittool.ClonedRepo{}, m.cloneRepoErr
	}
	return m.cloneRepoResult, nil
}

// MockWorkspace implements workspace.Workspace for testing
type MockWorkspace struct {
	readFileFromRepoCallCount int
	readFileFromRepoErr       error
	readFileFromRepoContent   map[string][]byte // key: filepath
	clearErr                  error
	createRepoFolderErr       error
}

func (m *MockWorkspace) Clear() error {
	return m.clearErr
}

func (m *MockWorkspace) CreateRepoFolder(repoURL string) (string, error) {
	return "", m.createRepoFolderErr
}

func (m *MockWorkspace) SaveSbom(models.Sbom) error {
	return nil
}

func (m *MockWorkspace) ReadFileFromRepo(repoPath string, relativePath string) ([]byte, error) {
	m.readFileFromRepoCallCount++
	if m.readFileFromRepoErr != nil {
		return nil, m.readFileFromRepoErr
	}
	if m.readFileFromRepoContent != nil {
		if content, ok := m.readFileFromRepoContent[relativePath]; ok {
			return content, nil
		}
	}
	return nil, errors.New("file not found")
}

func TestNewSyncer(t *testing.T) {
	mockCloner := &MockCloner{}
	mockWorkspace := &MockWorkspace{}

	syncer := NewSyncer(mockCloner, mockWorkspace)

	if syncer == nil {
		t.Error("NewSyncer returned nil")
	}

	if syncer.gitTool != mockCloner {
		t.Error("gitTool not set correctly")
	}

	if syncer.workspace != mockWorkspace {
		t.Error("workspace not set correctly")
	}

	if syncer.state.GitOpsRepos == nil {
		t.Error("GitOpsRepos map not initialized")
	}

	if len(syncer.state.GitOpsRepos) != 0 {
		t.Error("GitOpsRepos should be empty initially")
	}
}

func TestSyncer_Init_Success(t *testing.T) {
	mockCloner := &MockCloner{
		cloneRepoResult: gittool.ClonedRepo{
			Path:    "/tmp/repo1",
			RepoUrl: "https://github.com/example/repo1.git",
		},
	}
	mockWorkspace := &MockWorkspace{
		readFileFromRepoContent: map[string][]byte{
			"app/version.yaml": []byte("version: 1.0.0"),
		},
	}

	syncer := NewSyncer(mockCloner, mockWorkspace)

	gitOpsRepos := []config.GitOpsRepo{
		{
			Url: "https://github.com/example/repo1.git",
			GitOpsApplications: []config.GitOpsApplication{
				{
					ApplicationName: "app1",
					VersionIdentifiers: []config.VersionIdentifier{
						{
							Environment: "prod",
							Filepath:    "app/version.yaml",
							YamlPath:    ".version",
						},
					},
				},
			},
		},
	}

	err := syncer.Init(gitOpsRepos)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if len(syncer.state.GitOpsRepos) != 1 {
		t.Errorf("Expected 1 repo in state, got %d", len(syncer.state.GitOpsRepos))
	}

	repoState, ok := syncer.state.GitOpsRepos["https://github.com/example/repo1.git"]
	if !ok {
		t.Error("Expected repo not found in state")
	}

	if repoState.Repo.RepoUrl != "https://github.com/example/repo1.git" {
		t.Errorf("Expected RepoUrl to be set correctly, got %q", repoState.Repo.RepoUrl)
	}

	if len(repoState.AppStates) != 1 {
		t.Errorf("Expected 1 app state, got %d", len(repoState.AppStates))
	}

	if mockCloner.cloneRepoCallCount != 1 {
		t.Errorf("Expected CloneRepo to be called once, was called %d times", mockCloner.cloneRepoCallCount)
	}
}

func TestSyncer_Init_MultipleRepos(t *testing.T) {
	mockCloner := &MockCloner{
		cloneRepoResult: gittool.ClonedRepo{
			Path:    "/tmp/repo",
			RepoUrl: "https://github.com/example/repo.git",
		},
	}
	mockWorkspace := &MockWorkspace{
		readFileFromRepoContent: map[string][]byte{
			"version.yaml": []byte("version: 1.0.0"),
		},
	}

	syncer := NewSyncer(mockCloner, mockWorkspace)

	gitOpsRepos := []config.GitOpsRepo{
		{
			Url: "https://github.com/example/repo1.git",
			GitOpsApplications: []config.GitOpsApplication{
				{
					ApplicationName: "app1",
					VersionIdentifiers: []config.VersionIdentifier{
						{
							Environment: "prod",
							Filepath:    "version.yaml",
							YamlPath:    ".version",
						},
					},
				},
			},
		},
		{
			Url: "https://github.com/example/repo2.git",
			GitOpsApplications: []config.GitOpsApplication{
				{
					ApplicationName: "app2",
					VersionIdentifiers: []config.VersionIdentifier{
						{
							Environment: "staging",
							Filepath:    "version.yaml",
							YamlPath:    ".version",
						},
					},
				},
			},
		},
	}

	err := syncer.Init(gitOpsRepos)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if len(syncer.state.GitOpsRepos) != 2 {
		t.Errorf("Expected 2 repos in state, got %d", len(syncer.state.GitOpsRepos))
	}

	if mockCloner.cloneRepoCallCount != 2 {
		t.Errorf("Expected CloneRepo to be called twice, was called %d times", mockCloner.cloneRepoCallCount)
	}
}

func TestSyncer_Init_CloneError(t *testing.T) {
	cloneErr := errors.New("clone failed")
	mockCloner := &MockCloner{
		cloneRepoErr: cloneErr,
	}
	mockWorkspace := &MockWorkspace{}

	syncer := NewSyncer(mockCloner, mockWorkspace)

	gitOpsRepos := []config.GitOpsRepo{
		{
			Url: "https://github.com/example/repo.git",
			GitOpsApplications: []config.GitOpsApplication{
				{
					ApplicationName: "app1",
					VersionIdentifiers: []config.VersionIdentifier{
						{
							Environment: "prod",
							Filepath:    "version.yaml",
							YamlPath:    ".version",
						},
					},
				},
			},
		},
	}

	err := syncer.Init(gitOpsRepos)
	if err == nil {
		t.Error("Expected error from Init")
	}

	if !errors.Is(err, cloneErr) {
		t.Errorf("Expected error to be related to clone, got: %v", err)
	}

	if len(syncer.state.GitOpsRepos) != 0 {
		t.Errorf("Expected no repos in state after error, got %d", len(syncer.state.GitOpsRepos))
	}
}

func TestSyncer_Init_ReadFileError(t *testing.T) {
	mockCloner := &MockCloner{
		cloneRepoResult: gittool.ClonedRepo{
			Path:    "/tmp/repo",
			RepoUrl: "https://github.com/example/repo.git",
		},
	}
	readFileErr := errors.New("file read failed")
	mockWorkspace := &MockWorkspace{
		readFileFromRepoErr: readFileErr,
	}

	syncer := NewSyncer(mockCloner, mockWorkspace)

	gitOpsRepos := []config.GitOpsRepo{
		{
			Url: "https://github.com/example/repo.git",
			GitOpsApplications: []config.GitOpsApplication{
				{
					ApplicationName: "app1",
					VersionIdentifiers: []config.VersionIdentifier{
						{
							Environment: "prod",
							Filepath:    "version.yaml",
							YamlPath:    ".version",
						},
					},
				},
			},
		},
	}

	err := syncer.Init(gitOpsRepos)
	if err == nil {
		t.Error("Expected error from Init")
	}

	if !errors.Is(err, readFileErr) {
		t.Errorf("Expected error to be related to file read, got: %v", err)
	}
}

func TestSyncer_Init_ExtractValueError(t *testing.T) {
	mockCloner := &MockCloner{
		cloneRepoResult: gittool.ClonedRepo{
			Path:    "/tmp/repo",
			RepoUrl: "https://github.com/example/repo.git",
		},
	}
	mockWorkspace := &MockWorkspace{
		readFileFromRepoContent: map[string][]byte{
			"version.yaml": []byte("invalid: yaml: :"),
		},
	}

	syncer := NewSyncer(mockCloner, mockWorkspace)

	gitOpsRepos := []config.GitOpsRepo{
		{
			Url: "https://github.com/example/repo.git",
			GitOpsApplications: []config.GitOpsApplication{
				{
					ApplicationName: "app1",
					VersionIdentifiers: []config.VersionIdentifier{
						{
							Environment: "prod",
							Filepath:    "version.yaml",
							YamlPath:    ".nonexistent",
						},
					},
				},
			},
		},
	}

	err := syncer.Init(gitOpsRepos)
	if err == nil {
		t.Error("Expected error from Init")
	}
}

func TestSyncer_Init_MultipleAppsPerRepo(t *testing.T) {
	mockCloner := &MockCloner{
		cloneRepoResult: gittool.ClonedRepo{
			Path:    "/tmp/repo",
			RepoUrl: "https://github.com/example/repo.git",
		},
	}
	mockWorkspace := &MockWorkspace{
		readFileFromRepoContent: map[string][]byte{
			"app1/version.yaml": []byte("version: 1.0.0"),
			"app2/version.yaml": []byte("version: 2.0.0"),
		},
	}

	syncer := NewSyncer(mockCloner, mockWorkspace)

	gitOpsRepos := []config.GitOpsRepo{
		{
			Url: "https://github.com/example/repo.git",
			GitOpsApplications: []config.GitOpsApplication{
				{
					ApplicationName: "app1",
					VersionIdentifiers: []config.VersionIdentifier{
						{
							Environment: "prod",
							Filepath:    "app1/version.yaml",
							YamlPath:    ".version",
						},
						{
							Environment: "staging",
							Filepath:    "app1/version.yaml",
							YamlPath:    ".version",
						},
					},
				},
				{
					ApplicationName: "app2",
					VersionIdentifiers: []config.VersionIdentifier{
						{
							Environment: "prod",
							Filepath:    "app2/version.yaml",
							YamlPath:    ".version",
						},
					},
				},
			},
		},
	}

	err := syncer.Init(gitOpsRepos)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	repoState, ok := syncer.state.GitOpsRepos["https://github.com/example/repo.git"]
	if !ok {
		t.Fatal("Expected repo not found in state")
	}

	// Should have 3 app states: app1-prod, app1-staging, app2-prod
	if len(repoState.AppStates) != 3 {
		t.Errorf("Expected 3 app states, got %d", len(repoState.AppStates))
	}

	// Check specific app states
	key1 := AppStateKey{AppName: "app1", Environment: "prod"}
	state1, ok := repoState.AppStates[key1]
	if !ok {
		t.Error("Expected app1-prod state")
	} else if state1.CurrentVersion != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", state1.CurrentVersion)
	}

	key2 := AppStateKey{AppName: "app2", Environment: "prod"}
	state2, ok := repoState.AppStates[key2]
	if !ok {
		t.Error("Expected app2-prod state")
	} else if state2.CurrentVersion != "2.0.0" {
		t.Errorf("Expected version 2.0.0, got %s", state2.CurrentVersion)
	}
}

func TestSyncer_Init_EmptyRepoList(t *testing.T) {
	mockCloner := &MockCloner{}
	mockWorkspace := &MockWorkspace{}

	syncer := NewSyncer(mockCloner, mockWorkspace)

	err := syncer.Init([]config.GitOpsRepo{})
	if err != nil {
		t.Fatalf("Init should not fail with empty list: %v", err)
	}

	if len(syncer.state.GitOpsRepos) != 0 {
		t.Errorf("Expected no repos in state, got %d", len(syncer.state.GitOpsRepos))
	}
}

func TestSyncer_Init_AppStateProperties(t *testing.T) {
	mockCloner := &MockCloner{
		cloneRepoResult: gittool.ClonedRepo{
			Path:    "/tmp/repo",
			RepoUrl: "https://github.com/example/repo.git",
		},
	}
	mockWorkspace := &MockWorkspace{
		readFileFromRepoContent: map[string][]byte{
			"version.yaml": []byte("version: 2.5.3"),
		},
	}

	syncer := NewSyncer(mockCloner, mockWorkspace)

	gitOpsRepos := []config.GitOpsRepo{
		{
			Url: "https://github.com/example/repo.git",
			GitOpsApplications: []config.GitOpsApplication{
				{
					ApplicationName: "testapp",
					VersionIdentifiers: []config.VersionIdentifier{
						{
							Environment: "staging",
							Filepath:    "version.yaml",
							YamlPath:    ".version",
						},
					},
				},
			},
		},
	}

	err := syncer.Init(gitOpsRepos)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	repoState := syncer.state.GitOpsRepos["https://github.com/example/repo.git"]
	appState := repoState.AppStates[AppStateKey{AppName: "testapp", Environment: "staging"}]

	if appState.AppName != "testapp" {
		t.Errorf("Expected AppName 'testapp', got %q", appState.AppName)
	}

	if appState.Environment != "staging" {
		t.Errorf("Expected Environment 'staging', got %q", appState.Environment)
	}

	if appState.CurrentVersion != "2.5.3" {
		t.Errorf("Expected CurrentVersion '2.5.3', got %q", appState.CurrentVersion)
	}

	if appState.Handled != false {
		t.Error("Expected Handled to be false initially")
	}
}
