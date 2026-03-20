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
	state          SyncState
	gitTool        gittool.Cloner
	workspace      workspace.Workspace
	valueExtractor query.ValueExtractor
}

// NewSyncer creates a new instance of Syncer with the provided git tool and workspace
func NewSyncer(gitTool gittool.Cloner, workspace workspace.Workspace) *Syncer {
	return &Syncer{
		state: SyncState{
			GitOpsRepos: make(map[string]*GitOpsRepoState),
		},
		gitTool:        gitTool,
		workspace:      workspace,
		valueExtractor: query.NewYqValueExtractor(),
	}
}

func (s *Syncer) Init(gitOpsRepos []config.GitOpsRepo) error {
	s.state.GitOpsRepos = make(map[string]*GitOpsRepoState)

	for _, repo := range gitOpsRepos {
		repoState, err := s.initGitOpsRepo(repo)
		if err != nil {
			slog.Error("Failed to initialize GitOps repo", "repoUrl", repo.Url, "error", err)
			return fmt.Errorf("failed to initialize GitOps repo %s: %w", repo.Url, err)
		}
		s.state.GitOpsRepos[repo.Url] = &repoState
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

	appStates := make(map[AppStateKey]*GitOpsAppState)
	extractor := query.NewYqValueExtractor()

	for _, app := range gitOpsRepo.GitOpsApplications {
		for _, versionIdentifier := range app.VersionIdentifiers {

			internalVersionIdentfier := mapToInternalVersionIdentifer(versionIdentifier)

			versionIdentifierFile, err := s.workspace.ReadFileFromRepo(clonedRepo.Path, internalVersionIdentfier.filePath)
			if err != nil {
				slog.Error("Failed to read version file",
					"app", app.ApplicationName,
					"environment", versionIdentifier.Environment,
					"filepath", versionIdentifier.Filepath,
					"error", err)
				return GitOpsRepoState{}, fmt.Errorf("failed to read file %s: %w", versionIdentifier.Filepath, err)
			}

			version, err := extractor.ExtractValue(versionIdentifierFile, versionIdentifier.YamlPath)
			if err != nil {
				slog.Error("Failed to extract version from file",
					"app", app.ApplicationName,
					"environment", versionIdentifier.Environment,
					"filepath", versionIdentifier.Filepath,
					"yamlPath", versionIdentifier.YamlPath,
					"error", err)
				return GitOpsRepoState{}, fmt.Errorf("failed to extract version: %w", err)
			}

			appState := GitOpsAppState{
				AppName:           app.ApplicationName,
				VersionIdentifier: mapToInternalVersionIdentifer(versionIdentifier),
				CurrentVersion:    version,
				Handled:           false,
			}
			appStateKey := AppStateKey{
				AppName:     app.ApplicationName,
				Environment: versionIdentifier.Environment,
			}
			appStates[appStateKey] = &appState

			slog.Info("Extracted app version",
				"app", app.ApplicationName,
				"environment", versionIdentifier.Environment,
				"version", version)
		}
	}

	return GitOpsRepoState{Repo: clonedRepo, AppStates: appStates}, nil
}

func mapToInternalVersionIdentifer(configIdentifier config.VersionIdentifier) VersionIdentifier {
	return VersionIdentifier{
		env:      configIdentifier.Environment,
		filePath: configIdentifier.Filepath,
		yamlPath: configIdentifier.YamlPath,
	}
}

func (s *Syncer) Reconcile() {
	for _, repoState := range s.state.GitOpsRepos {
		err := s.reconcileGitOpsRepo(repoState)
		if err != nil {
			slog.Warn("Error reconciling GitOps repo, skipping to next", "repo", repoState.Repo.RepoUrl, "error", err)
		}
	}

}

func (s *Syncer) reconcileGitOpsRepo(repoState *GitOpsRepoState) error {
	updated, err := repoState.Repo.UpdateIfAvailable()

	if err != nil {
		slog.Warn("Error reconciling GitOps repo", "repo", repoState.Repo.RepoUrl, "error", err)
		return err
	}

	if updated {
		for _, appstate := range repoState.AppStates {
			maybeNewVersion, err := s.getVersionForApp(repoState.Repo, appstate)
			if err != nil {
				return err
			}

			if appstate.CurrentVersion != maybeNewVersion {
				appstate.CurrentVersion = maybeNewVersion
				appstate.Handled = false
			}
		}
	}

	s.checkUnhandledChanges(repoState)

	return nil
}

// Checks for unhandled changes and calls the registered handler
func (s *Syncer) checkUnhandledChanges(repoState *GitOpsRepoState) {
	for _, appState := range repoState.AppStates {

		if appState.Handled {
			continue
		}
		slog.Info("App changed, handle new version", "app", appState.AppName, "env", appState.VersionIdentifier.env, "version", appState.CurrentVersion)
	}
}

// Gets the version of an appstate based on its VersionIdentifier
func (s *Syncer) getVersionForApp(gitopsRepo gittool.ClonedRepo, app *GitOpsAppState) (string, error) {

	versionIdentifierFile, err := s.workspace.ReadFileFromRepo(gitopsRepo.Path, app.VersionIdentifier.filePath)
	if err != nil {
		slog.Error("Failed to read version file",
			"app", app.AppName,
			"environment", app.VersionIdentifier.env,
			"filepath", app.VersionIdentifier.filePath,
			"error", err)
		return "", err
	}

	version, err := s.valueExtractor.ExtractValue(versionIdentifierFile, app.VersionIdentifier.yamlPath)
	if err != nil {
		slog.Error("Failed to extract version from file",
			"app", app.AppName,
			"environment", app.VersionIdentifier.env,
			"filepath", app.VersionIdentifier.filePath,
			"yamlPath", app.VersionIdentifier.yamlPath,
			"error", err)
		return "", fmt.Errorf("failed to extract version: %w", err)
	}
	return version, nil

}
