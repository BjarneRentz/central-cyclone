package gitops

import (
	"central-cyclone/internal/config"
	"central-cyclone/internal/gittool"
	"central-cyclone/internal/query"
	"central-cyclone/internal/workspace"
	"fmt"
	"log/slog"
)

type Syncer struct {
	state     SyncState
	gitTool   gittool.Cloner
	workspace workspace.Workspace
}

// NewSyncer creates a new instance of Syncer with the provided git tool and workspace
func NewSyncer(gitTool gittool.Cloner, workspace workspace.Workspace) *Syncer {
	return &Syncer{
		state: SyncState{
			GitOpsRepos: make(map[string]GitOpsRepoState),
		},
		gitTool:   gitTool,
		workspace: workspace,
	}
}

func (s *Syncer) Init(gitOpsRepos []config.GitOpsRepo) error {
	s.state.GitOpsRepos = make(map[string]GitOpsRepoState)

	for _, repo := range gitOpsRepos {
		repoState, err := s.initGitOpsRepo(repo)
		if err != nil {
			slog.Error("Failed to initialize GitOps repo", "repoUrl", repo.Url, "error", err)
			return fmt.Errorf("failed to initialize GitOps repo %s: %w", repo.Url, err)
		}
		s.state.GitOpsRepos[repo.Url] = repoState
		slog.Info("Successfully initialized GitOps repo", "repoUrl", repo.Url, "apps", len(repoState.AppStates))
	}

	return nil
}

func (s *Syncer) initGitOpsRepo(gitOpsRepo config.GitOpsRepo) (GitOpsRepoState, error) {
	clonedRepo, err := s.gitTool.CloneRepo(gitOpsRepo.Url)
	if err != nil {
		slog.Error("Could not clone GitOpsRepo", "repoUrl", gitOpsRepo.Url, "error", err)
		return GitOpsRepoState{}, err
	}

	appStates := make(map[AppStateKey]AppState)
	extractor := query.NewYqValueExtractor()

	for _, app := range gitOpsRepo.GitOpsApplications {
		for _, versionIdentifier := range app.VersionIdentifiers {

			fileContent, err := s.workspace.ReadFileFromRepo(clonedRepo.Path, versionIdentifier.Filepath)
			if err != nil {
				slog.Error("Failed to read version file",
					"app", app.ApplicationName,
					"environment", versionIdentifier.Environment,
					"filepath", versionIdentifier.Filepath,
					"error", err)
				return GitOpsRepoState{}, fmt.Errorf("failed to read file %s: %w", versionIdentifier.Filepath, err)
			}

			version, err := extractor.ExtractValue(fileContent, versionIdentifier.YamlPath)
			if err != nil {
				slog.Error("Failed to extract version from file",
					"app", app.ApplicationName,
					"environment", versionIdentifier.Environment,
					"filepath", versionIdentifier.Filepath,
					"yamlPath", versionIdentifier.YamlPath,
					"error", err)
				return GitOpsRepoState{}, fmt.Errorf("failed to extract version: %w", err)
			}

			appState := AppState{
				AppName:        app.ApplicationName,
				Environment:    versionIdentifier.Environment,
				CurrentVersion: version,
				Handled:        false,
			}
			appStateKey := AppStateKey{
				AppName:     app.ApplicationName,
				Environment: versionIdentifier.Environment,
			}
			appStates[appStateKey] = appState

			slog.Info("Extracted app version",
				"app", app.ApplicationName,
				"environment", versionIdentifier.Environment,
				"version", version)
		}
	}

	return GitOpsRepoState{Repo: clonedRepo, AppStates: appStates}, nil
}
