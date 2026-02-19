package gitops

import (
	"central-cyclone/internal/config"
	"central-cyclone/internal/gittool"
	"central-cyclone/internal/models"
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
)

// tempWorkspace is a minimal Workspace implementation for integration tests
type tempWorkspace struct {
	root string
}

func newTempWorkspace(t *testing.T) *tempWorkspace {
	return &tempWorkspace{root: t.TempDir()}
}

func (w *tempWorkspace) Clear() error                 { return nil }
func (w *tempWorkspace) SaveSbom(s models.Sbom) error { return nil }

func (w *tempWorkspace) CreateRepoFolder(repoURL string) (string, error) {
	// create a folder under root derived from repoURL base
	name := filepath.Base(repoURL)
	if name == "." || name == "/" || name == "" {
		name = "repo"
	}
	target := filepath.Join(w.root, name+"-cloned")
	if err := os.MkdirAll(target, 0o755); err != nil {
		return "", err
	}
	return target, nil
}

func (w *tempWorkspace) ReadFileFromRepo(repoPath string, relativePath string) ([]byte, error) {
	p := filepath.Join(repoPath, relativePath)
	return os.ReadFile(p)
}

func TestSyncer_FullChain_LocalRepoClone(t *testing.T) {
	// 1) create a source git repo with the YAML file
	source := t.TempDir()
	r, err := git.PlainInit(source, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}

	// write file
	filePath := filepath.Join(source, "app", "version.yaml")
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filePath, []byte("version: 9.9.9\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	wt, err := r.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}
	if _, err := wt.Add("app/version.yaml"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, err := wt.Commit("add version", &git.CommitOptions{
		Author: &object.Signature{Name: "Tester", Email: "t@example.com", When: time.Now()},
	}); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// 2) prepare temp workspace and local cloner that will clone from 'source'
	tw := newTempWorkspace(t)
	cloner := gittool.CreateLocalGitCloner(tw)

	// 3) create syncer using real cloner and temp workspace
	s := NewSyncer(cloner, tw)

	// 4) prepare config to point at the local repo (use the source path as URL)
	g := config.GitOpsRepo{
		Url: source,
		GitOpsApplications: []config.GitOpsApplication{
			{
				ApplicationName: "app",
				VersionIdentifiers: []config.VersionIdentifier{
					{
						Environment: "prod",
						Filepath:    "app/version.yaml",
						YamlPath:    ".version",
					},
				},
			},
		},
	}

	// 5) run Init which will clone and extract
	if err := s.Init([]config.GitOpsRepo{g}); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 6) assert the state contains the expected version
	rs, ok := s.state.GitOpsRepos[source]
	if !ok {
		t.Fatalf("repo state not found")
	}
	if rs.Repo.RepoUrl != source {
		t.Fatalf("expected repo url %q got %q", source, rs.Repo.RepoUrl)
	}
	key := AppStateKey{AppName: "app", Environment: "prod"}
	as, ok := rs.AppStates[key]
	if !ok {
		t.Fatalf("app state missing")
	}
	if as.CurrentVersion != "9.9.9" {
		t.Fatalf("extracted version mismatch: %q", as.CurrentVersion)
	}
}
