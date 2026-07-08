package config

import (
	"testing"
)

func TestNewConfigProvider_Validation_MissingApplication(t *testing.T) {
	settings := &Settings{
		GitOpsRepos: []GitOpsRepo{
			{
				Url: "https://github.com/test/repo.git",
				GitOpsApplications: []GitOpsApplication{
					{
						ApplicationName: "test-app",
						VersionIdentifiers: []VersionIdentifier{
							{Environment: "prod", Filepath: "test.yaml", YamlPath: "test"},
						},
					},
				},
			},
		},
	}

	_, err := NewConfigProvider(settings)
	if err == nil {
		t.Fatal("expected validation error for missing Application entry")
	}
	if err.Error() != "GitOps application 'test-app' has no corresponding Application entry in config" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestNewConfigProvider_Validation_MissingApplicationRepo(t *testing.T) {
	settings := &Settings{
		Applications: []Application{
			{Name: "test-app", Type: "go"},
		},
		GitOpsRepos: []GitOpsRepo{
			{
				Url: "https://github.com/test/repo.git",
				GitOpsApplications: []GitOpsApplication{
					{
						ApplicationName: "test-app",
						VersionIdentifiers: []VersionIdentifier{
							{Environment: "prod", Filepath: "test.yaml", YamlPath: "test"},
						},
					},
				},
			},
		},
	}

	_, err := NewConfigProvider(settings)
	if err == nil {
		t.Fatal("expected validation error for missing ApplicationRepo entry")
	}
	if err.Error() != "GitOps application 'test-app' has no corresponding ApplicationRepo entry in config" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestNewConfigProvider_Validation_MissingProject(t *testing.T) {
	projectId := "test-project-id"
	settings := &Settings{
		Applications: []Application{
			{
				Name: "test-app",
				Type: "go",
				Projects: []Project{
					{Name: "test-app", Environment: "staging", ProjectId: &projectId},
				},
			},
		},
		ApplicationRepos: []ApplicationRepo{
			{Applications: []string{"test-app"}, RepoUrl: "https://github.com/test/repo.git"},
		},
		GitOpsRepos: []GitOpsRepo{
			{
				Url: "https://github.com/test/gitops.git",
				GitOpsApplications: []GitOpsApplication{
					{
						ApplicationName: "test-app",
						VersionIdentifiers: []VersionIdentifier{
							{Environment: "prod", Filepath: "test.yaml", YamlPath: "test"},
						},
					},
				},
			},
		},
	}

	_, err := NewConfigProvider(settings)
	if err == nil {
		t.Fatal("expected validation error for missing Project with environment 'prod'")
	}
	expectedError := "GitOps application 'test-app' with environment 'prod' has no matching Project in Application 'test-app' projects"
	if err.Error() != expectedError {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestNewConfigProvider_Validation_Success(t *testing.T) {
	projectId := "test-project-id"
	settings := &Settings{
		Applications: []Application{
			{
				Name: "test-app",
				Type: "go",
				Projects: []Project{
					{Name: "test-app", Environment: "prod", ProjectId: &projectId},
				},
			},
		},
		ApplicationRepos: []ApplicationRepo{
			{Applications: []string{"test-app"}, RepoUrl: "https://github.com/test/repo.git"},
		},
		GitOpsRepos: []GitOpsRepo{
			{
				Url: "https://github.com/test/gitops.git",
				GitOpsApplications: []GitOpsApplication{
					{
						ApplicationName: "test-app",
						VersionIdentifiers: []VersionIdentifier{
							{Environment: "prod", Filepath: "test.yaml", YamlPath: "test"},
						},
					},
				},
			},
		},
	}

	provider, err := NewConfigProvider(settings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestNewConfigProvider_Validation_MultipleApps(t *testing.T) {
	projectId1 := "project-1"
	projectId2 := "project-2"
	settings := &Settings{
		Applications: []Application{
			{
				Name: "app1",
				Type: "go",
				Projects: []Project{
					{Name: "app1", Environment: "prod", ProjectId: &projectId1},
					{Name: "app1", Environment: "staging", ProjectId: &projectId2},
				},
			},
			{
				Name: "app2",
				Type: "node",
				Projects: []Project{
					{Name: "app2", Environment: "prod", ProjectId: &projectId1},
				},
			},
		},
		ApplicationRepos: []ApplicationRepo{
			{Applications: []string{"app1", "app2"}, RepoUrl: "https://github.com/test/repo.git"},
		},
		GitOpsRepos: []GitOpsRepo{
			{
				Url: "https://github.com/test/gitops.git",
				GitOpsApplications: []GitOpsApplication{
					{
						ApplicationName: "app1",
						VersionIdentifiers: []VersionIdentifier{
							{Environment: "prod", Filepath: "test.yaml", YamlPath: "test"},
							{Environment: "staging", Filepath: "test.yaml", YamlPath: "test"},
						},
					},
					{
						ApplicationName: "app2",
						VersionIdentifiers: []VersionIdentifier{
							{Environment: "prod", Filepath: "test.yaml", YamlPath: "test"},
						},
					},
				},
			},
		},
	}

	provider, err := NewConfigProvider(settings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
}
