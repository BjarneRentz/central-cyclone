package gittool

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

type testRepoFixture struct {
	path           string
	firstCommit    string
	secondCommit   string
	lightweightTag string
	annotatedTag   string
}

func newTestRepo(t *testing.T, lightweightTag, annotatedTag string) testRepoFixture {
	t.Helper()
	tmpDir := t.TempDir()

	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	writeAndCommit := func(name, content, message string) string {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
		if _, err := wt.Add(name); err != nil {
			t.Fatalf("failed to add file: %v", err)
		}
		hash, err := wt.Commit(message, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "test",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})
		if err != nil {
			t.Fatalf("failed to commit: %v", err)
		}
		return hash.String()
	}

	firstCommit := writeAndCommit("README.md", "hello world\n", "initial commit")
	if _, err := repo.CreateTag(lightweightTag, plumbing.NewHash(firstCommit), nil); err != nil {
		t.Fatalf("failed to create lightweight tag: %v", err)
	}

	secondCommit := writeAndCommit("CHANGES.md", "more changes\n", "second commit")
	if _, err := repo.CreateTag(annotatedTag, plumbing.NewHash(secondCommit), &git.CreateTagOptions{
		Tagger: &object.Signature{
			Name:  "test",
			Email: "test@example.com",
			When:  time.Now(),
		},
		Message: "release " + annotatedTag,
	}); err != nil {
		t.Fatalf("failed to create annotated tag: %v", err)
	}

	return testRepoFixture{
		path:           tmpDir,
		firstCommit:    firstCommit,
		secondCommit:   secondCommit,
		lightweightTag: lightweightTag,
		annotatedTag:   annotatedTag,
	}
}

func TestClonedRepo_CheckoutRevision_FullSHA(t *testing.T) {
	fixture := newTestRepo(t, "v1.0.0", "v2.0.0")
	repo := &ClonedRepo{Path: fixture.path}

	if err := repo.CheckoutRevision(fixture.firstCommit); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	current, err := repo.GetCurrentRevision()
	if err != nil {
		t.Fatalf("failed to get current revision: %v", err)
	}
	if current != fixture.firstCommit {
		t.Fatalf("expected current revision %q, got %q", fixture.firstCommit, current)
	}
}

func TestClonedRepo_CheckoutRevision_PlainTag(t *testing.T) {
	fixture := newTestRepo(t, "v1.0.0", "v2.0.0")
	repo := &ClonedRepo{Path: fixture.path}

	if err := repo.CheckoutRevision(fixture.lightweightTag); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	current, err := repo.GetCurrentRevision()
	if err != nil {
		t.Fatalf("failed to get current revision: %v", err)
	}
	if current != fixture.firstCommit {
		t.Fatalf("expected current revision %q, got %q", fixture.firstCommit, current)
	}
}

func TestClonedRepo_CheckoutRevision_AnnotatedTag(t *testing.T) {
	fixture := newTestRepo(t, "v1.0.0", "v2.0.0")
	repo := &ClonedRepo{Path: fixture.path}

	if err := repo.CheckoutRevision(fixture.annotatedTag); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	current, err := repo.GetCurrentRevision()
	if err != nil {
		t.Fatalf("failed to get current revision: %v", err)
	}
	if current != fixture.secondCommit {
		t.Fatalf("expected current revision %q, got %q", fixture.secondCommit, current)
	}
}

func TestClonedRepo_CheckoutRevision_TagPlusShortHex(t *testing.T) {
	fixture := newTestRepo(t, "v1.0.0", "v2.0.0")
	repo := &ClonedRepo{Path: fixture.path}

	shortHex := fixture.firstCommit[:8]
	revision := fixture.annotatedTag + "+" + shortHex

	if err := repo.CheckoutRevision(revision); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	current, err := repo.GetCurrentRevision()
	if err != nil {
		t.Fatalf("failed to get current revision: %v", err)
	}
	if current != fixture.firstCommit {
		t.Fatalf("expected current revision %q, got %q", fixture.firstCommit, current)
	}
}

func TestClonedRepo_CheckoutRevision_TagPlusFullHex(t *testing.T) {
	fixture := newTestRepo(t, "v1.0.0", "v2.0.0")
	repo := &ClonedRepo{Path: fixture.path}

	revision := fixture.lightweightTag + "+" + fixture.secondCommit

	if err := repo.CheckoutRevision(revision); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	current, err := repo.GetCurrentRevision()
	if err != nil {
		t.Fatalf("failed to get current revision: %v", err)
	}
	if current != fixture.secondCommit {
		t.Fatalf("expected current revision %q, got %q", fixture.secondCommit, current)
	}
}

func TestClonedRepo_CheckoutRevision_TagPlusNonHexSuffix(t *testing.T) {
	fixture := newTestRepo(t, "v1.0.0", "v2.0.0")
	repo := &ClonedRepo{Path: fixture.path}

	revision := fixture.lightweightTag + "+34asdadasd"

	err := repo.CheckoutRevision(revision)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), revision) {
		t.Fatalf("expected error to contain full revision %q, got: %v", revision, err)
	}
}

func TestClonedRepo_CheckoutRevision_TagPlusShortSuffix(t *testing.T) {
	fixture := newTestRepo(t, "v1.0.0", "v2.0.0")
	repo := &ClonedRepo{Path: fixture.path}

	revision := fixture.lightweightTag + "+ab1"

	err := repo.CheckoutRevision(revision)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), revision) {
		t.Fatalf("expected error to contain full revision %q, got: %v", revision, err)
	}
}

func TestClonedRepo_CheckoutRevision_TagPlusEmptySuffix(t *testing.T) {
	fixture := newTestRepo(t, "v1.0.0", "v2.0.0")
	repo := &ClonedRepo{Path: fixture.path}

	revision := fixture.lightweightTag + "+"

	err := repo.CheckoutRevision(revision)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), revision) {
		t.Fatalf("expected error to contain full revision %q, got: %v", revision, err)
	}
}

func TestClonedRepo_CheckoutRevision_TagPlusUnknownHex(t *testing.T) {
	fixture := newTestRepo(t, "v1.0.0", "v2.0.0")
	repo := &ClonedRepo{Path: fixture.path}

	revision := fixture.lightweightTag + "+deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"

	err := repo.CheckoutRevision(revision)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), revision) {
		t.Fatalf("expected error to contain full revision %q, got: %v", revision, err)
	}
}
