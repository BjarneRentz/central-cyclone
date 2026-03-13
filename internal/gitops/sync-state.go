package gitops

import (
	"central-cyclone/internal/gittool"
)

type SyncState struct {
	GitOpsRepos map[string]GitOpsRepoState // Key: Repo URL
}

type GitOpsRepoState struct {
	Repo      gittool.ClonedRepo
	AppStates map[AppStateKey]GitOpsAppState
}

type GitOpsAppState struct {
	AppName           string
	VersionIdentifier VersionIdentifier
	CurrentVersion    string
	Handled           bool
}

type VersionIdentifier struct {
	env      string
	filePath string
	yamlPath string
}

type AppStateKey struct {
	AppName     string
	Environment string
}
