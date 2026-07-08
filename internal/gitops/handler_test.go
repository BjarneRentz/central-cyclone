package gitops

import (
	"central-cyclone/internal/analyzer"
	"central-cyclone/internal/config"
	"central-cyclone/internal/gittool"
	"central-cyclone/internal/models"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
)

type MockRepoCloner struct {
	repo        gittool.ClonedRepo
	err         error
	receivedURL string
	called      bool
}

func (m *MockRepoCloner) CloneRepo(repoURL string) (gittool.ClonedRepo, error) {
	return m.CloneOrUpdateRepo(repoURL)
}

func (m *MockRepoCloner) CloneOrUpdateRepo(repoURL string) (gittool.ClonedRepo, error) {
	m.called = true
	m.receivedURL = repoURL
	return m.repo, m.err
}

type MockAnalyzer struct {
	receivedRepo   gittool.ClonedRepo
	receivedTarget *analyzer.ScanTarget
	result         models.Sbom
	err            error
	called         bool
}

func (m *MockAnalyzer) AnalyzeProject(repo gittool.ClonedRepo, target *analyzer.ScanTarget) (models.Sbom, error) {
	m.called = true
	m.receivedRepo = repo
	m.receivedTarget = target
	return m.result, m.err
}

type MockUploader struct {
	receivedSbom models.Sbom
	err          error
	called       bool
}

func (m *MockUploader) UploadSBOM(ctx context.Context, sbom models.Sbom) error {
	m.called = true
	m.receivedSbom = sbom
	return m.err
}

func createTempGitRepoWithTag(t *testing.T, tag string) string {
	t.Helper()
	tmpDir := t.TempDir()

	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	readmePath := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("hello world\n"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	if _, err := wt.Add("README.md"); err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	commitHash, err := wt.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	if _, err := repo.CreateTag(tag, commitHash, nil); err != nil {
		t.Fatalf("failed to create tag: %v", err)
	}

	return tmpDir
}

func stringPtr(value string) *string {
	return &value
}

func TestCreateSbomChangeHandler_HandleAppChange_Success(t *testing.T) {
	tag := "v1.0.0"
	tmpRepo := createTempGitRepoWithTag(t, tag)

	settings := &config.Settings{
		Applications: []config.Application{
			{
				Name:     "my-app",
				Type:     "go",
				RepoPath: stringPtr("src"),
				Projects: []config.Project{
					{
						Environment: "prod",
						ProjectId:   stringPtr("project-123"),
					},
				},
			},
		},
		ApplicationRepos: []config.ApplicationRepo{
			{
				Applications: []string{"my-app"},
				RepoUrl:      tmpRepo,
			},
		},
	}

	configProvider, err := config.NewConfigProvider(settings)
	if err != nil {
		t.Fatalf("failed to create config provider: %v", err)
	}
	mockCloner := &MockRepoCloner{
		repo: gittool.ClonedRepo{
			Path:    tmpRepo,
			RepoUrl: tmpRepo,
		},
	}
	mockAnalyzer := &MockAnalyzer{
		result: models.Sbom{
			ProjectId:   "project-123",
			ProjectType: "go",
			Data:        "sbom-content",
		},
	}
	mockUploader := &MockUploader{}

	handler := CreateSbomChangeHandler{
		configProvider:          configProvider,
		gitTool:                 mockCloner,
		sbomAnalyzer:            mockAnalyzer,
		dependencyTrackUploader: mockUploader,
	}

	if err := handler.HandleAppChange(context.TODO(), "my-app", "prod", tag); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if !mockCloner.called {
		t.Fatal("expected CloneOrUpdateRepo to be called")
	}

	if !mockAnalyzer.called {
		t.Fatal("expected AnalyzeProject to be called")
	}

	if !mockUploader.called {
		t.Fatal("expected UploadSBOM to be called")
	}

	if mockAnalyzer.receivedTarget == nil {
		t.Fatal("expected scan target to be passed to analyzer")
	}

	if mockAnalyzer.receivedTarget.ProjectId != "project-123" {
		t.Fatalf("unexpected project id: %s", mockAnalyzer.receivedTarget.ProjectId)
	}

	if mockUploader.receivedSbom.ProjectId != "project-123" {
		t.Fatalf("unexpected uploaded sbom project id: %s", mockUploader.receivedSbom.ProjectId)
	}
}

func TestCreateSbomChangeHandler_HandleAppChange_ConfigRepoError(t *testing.T) {
	settings := &config.Settings{
		ApplicationRepos: []config.ApplicationRepo{},
	}
	configProvider, err := config.NewConfigProvider(settings)
	if err != nil {
		t.Fatalf("failed to create config provider: %v", err)
	}

	handler := CreateSbomChangeHandler{
		configProvider:          configProvider,
		gitTool:                 &MockRepoCloner{},
		sbomAnalyzer:            &MockAnalyzer{},
		dependencyTrackUploader: &MockUploader{},
	}

	err = handler.HandleAppChange(context.TODO(), "missing-app", "prod", "v1.0.0")
	if err == nil {
		t.Fatal("expected an error when application repo is missing")
	}
}

func TestCreateSbomChangeHandler_HandleAppChange_CloneError(t *testing.T) {
	tag := "v1.0.0"
	tmpRepo := createTempGitRepoWithTag(t, tag)

	settings := &config.Settings{
		Applications: []config.Application{
			{
				Name:     "my-app",
				Type:     "go",
				RepoPath: stringPtr("src"),
				Projects: []config.Project{{Environment: "prod", ProjectId: stringPtr("project-123")}},
			},
		},
		ApplicationRepos: []config.ApplicationRepo{{Applications: []string{"my-app"}, RepoUrl: tmpRepo}},
	}
	configProvider, err := config.NewConfigProvider(settings)
	if err != nil {
		t.Fatalf("failed to create config provider: %v", err)
	}
	cloneErr := errors.New("clone failure")
	mockCloner := &MockRepoCloner{err: cloneErr}

	handler := CreateSbomChangeHandler{
		configProvider:          configProvider,
		gitTool:                 mockCloner,
		sbomAnalyzer:            &MockAnalyzer{},
		dependencyTrackUploader: &MockUploader{},
	}

	err = handler.HandleAppChange(context.TODO(), "my-app", "prod", tag)
	if !errors.Is(err, cloneErr) {
		t.Fatalf("expected clone error, got: %v", err)
	}
}

func TestCreateSbomChangeHandler_HandleAppChange_CheckoutTagError(t *testing.T) {
	tag := "v1.0.0"
	tmpRepo := createTempGitRepoWithTag(t, tag)

	settings := &config.Settings{
		Applications: []config.Application{
			{Name: "my-app", Type: "go", RepoPath: stringPtr("src"), Projects: []config.Project{{Environment: "prod", ProjectId: stringPtr("project-123")}}},
		},
		ApplicationRepos: []config.ApplicationRepo{{Applications: []string{"my-app"}, RepoUrl: tmpRepo}},
	}
	configProvider, err := config.NewConfigProvider(settings)
	if err != nil {
		t.Fatalf("failed to create config provider: %v", err)
	}
	mockCloner := &MockRepoCloner{
		repo: gittool.ClonedRepo{Path: tmpRepo, RepoUrl: tmpRepo},
	}

	handler := CreateSbomChangeHandler{
		configProvider:          configProvider,
		gitTool:                 mockCloner,
		sbomAnalyzer:            &MockAnalyzer{},
		dependencyTrackUploader: &MockUploader{},
	}

	err = handler.HandleAppChange(context.TODO(), "my-app", "prod", "missing-tag")
	if err == nil {
		t.Fatal("expected error when checkout tag fails")
	}
}

func TestCreateSbomChangeHandler_HandleAppChange_AnalysisError(t *testing.T) {
	tag := "v1.0.0"
	tmpRepo := createTempGitRepoWithTag(t, tag)

	settings := &config.Settings{
		Applications:     []config.Application{{Name: "my-app", Type: "go", RepoPath: stringPtr("src"), Projects: []config.Project{{Environment: "prod", ProjectId: stringPtr("project-123")}}}},
		ApplicationRepos: []config.ApplicationRepo{{Applications: []string{"my-app"}, RepoUrl: tmpRepo}},
	}
	configProvider, err := config.NewConfigProvider(settings)
	if err != nil {
		t.Fatalf("failed to create config provider: %v", err)
	}
	mockCloner := &MockRepoCloner{repo: gittool.ClonedRepo{Path: tmpRepo, RepoUrl: tmpRepo}}
	analysisErr := errors.New("analysis failed")
	mockAnalyzer := &MockAnalyzer{err: analysisErr}

	handler := CreateSbomChangeHandler{
		configProvider:          configProvider,
		gitTool:                 mockCloner,
		sbomAnalyzer:            mockAnalyzer,
		dependencyTrackUploader: &MockUploader{},
	}

	err = handler.HandleAppChange(context.TODO(), "my-app", "prod", tag)
	if !errors.Is(err, analysisErr) {
		t.Fatalf("expected analysis error, got: %v", err)
	}
}

func TestCreateSbomChangeHandler_HandleAppChange_UploadError(t *testing.T) {
	tag := "v1.0.0"
	tmpRepo := createTempGitRepoWithTag(t, tag)

	settings := &config.Settings{
		Applications:     []config.Application{{Name: "my-app", Type: "go", RepoPath: stringPtr("src"), Projects: []config.Project{{Environment: "prod", ProjectId: stringPtr("project-123")}}}},
		ApplicationRepos: []config.ApplicationRepo{{Applications: []string{"my-app"}, RepoUrl: tmpRepo}},
	}
	configProvider, err := config.NewConfigProvider(settings)
	if err != nil {
		t.Fatalf("failed to create config provider: %v", err)
	}
	mockCloner := &MockRepoCloner{repo: gittool.ClonedRepo{Path: tmpRepo, RepoUrl: tmpRepo}}
	mockAnalyzer := &MockAnalyzer{result: models.Sbom{ProjectId: "project-123", ProjectType: "go", Data: "sbom-data"}}
	uploadErr := errors.New("upload failed")
	mockUploader := &MockUploader{err: uploadErr}

	handler := CreateSbomChangeHandler{
		configProvider:          configProvider,
		gitTool:                 mockCloner,
		sbomAnalyzer:            mockAnalyzer,
		dependencyTrackUploader: mockUploader,
	}

	err = handler.HandleAppChange(context.TODO(), "my-app", "prod", tag)
	if !errors.Is(err, uploadErr) {
		t.Fatalf("expected upload error, got: %v", err)
	}
}
