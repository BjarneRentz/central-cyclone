package gitops

import "central-cyclone/internal/gittool"

type SyncState struct {
	GitOpsRepos map[string]GitOpsRepoState // Key: Repo URL
}

type GitOpsRepoState struct {
	Repo      gittool.ClonedRepo
	AppStates map[AppStateKey]AppState
}

type AppState struct {
	AppName        string
	Environment    string // Dev / Prod / etc.
	CurrentVersion string
	Handled        bool // Specified Version
}

type AppStateKey struct {
	AppName     string
	Environment string
}
