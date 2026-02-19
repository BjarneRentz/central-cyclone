package gitops

import (
	"central-cyclone/internal/config"
	"central-cyclone/internal/gittool"
	"central-cyclone/internal/workspace"
	"log/slog"
)

type Syncer struct {
	state     SyncState
	gitTool   gittool.Cloner
	workspace workspace.Workspace
}

func (s Syncer) Init(gitOpsRepos []config.GitOpsRepo) error {

	return nil
}

// Working Notes
// - Use Workspace to read files from a given repo? => Easier to mock for unit test?
// Gittool has worksapce via di, call gittool directly for cloning an repo
// Gittool can later also checkout a repo at a given hash / tag / etc

func (s Syncer) initGitOpsRepo(gitOpsRepo config.GitOpsRepo) (GitOpsRepoState, error) {
	clonedRepo, err := s.gitTool.CloneRepo(gitOpsRepo.Url)
	if err != nil {
		slog.Error("Could not clone GitOpsRepo", "repoUrl", gitOpsRepo.Url, "error", err)
		return GitOpsRepoState{}, err
	}

	var appStates map[AppStateKey]AppState = make(map[AppStateKey]AppState)

	for _, app := range gitOpsRepo.GitOpsApplications {
		for _, versionId := range app.VersionIdentifiers {
			// Read the file at versionId.Filepath in clonedRepo.Path
			// Parse the file (YAML/JSON) to get the version at versionId.YamlPath

			var appState AppState = AppState{
				AppName:        app.ApplicationName,
				Environment:    versionId.Environment,
				CurrentVersion: "1.0.0", // Placeholder, replace with actual version from file
			}
			var appStateKey AppStateKey = AppStateKey{AppName: app.ApplicationName, Environment: versionId.Environment}
			appStates[appStateKey] = appState

		}
	}

	// Get the App Versions and create an initial state

	return GitOpsRepoState{Repo: clonedRepo, AppStates: appStates}, nil
}
