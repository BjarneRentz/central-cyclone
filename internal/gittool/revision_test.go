package gittool

import (
	"strings"
	"testing"
)

func TestSplitBuildMetadataRevision(t *testing.T) {
	tests := []struct {
		name          string
		revision      string
		wantCommitish string
		wantOk        bool
	}{
		{
			name:          "no plus",
			revision:      "v1.0.0",
			wantCommitish: "",
			wantOk:        false,
		},
		{
			name:          "valid short hex suffix",
			revision:      "v1.0.0+34ab12cd",
			wantCommitish: "34ab12cd",
			wantOk:        true,
		},
		{
			name:          "valid full 40-char hex suffix",
			revision:      "v1.0.0+0123456789abcdef0123456789abcdef01234567",
			wantCommitish: "0123456789abcdef0123456789abcdef01234567",
			wantOk:        true,
		},
		{
			name:          "uppercase hex suffix",
			revision:      "v1.0.0+34AB12CD",
			wantCommitish: "34AB12CD",
			wantOk:        true,
		},
		{
			name:          "non-hex suffix",
			revision:      "v1.0.0+34asdadasd",
			wantCommitish: "",
			wantOk:        false,
		},
		{
			name:          "suffix shorter than minimum",
			revision:      "v1.0.0+ab1",
			wantCommitish: "",
			wantOk:        false,
		},
		{
			name:          "suffix longer than a full SHA-1",
			revision:      "v1.0.0+0123456789abcdef0123456789abcdef012345678",
			wantCommitish: "",
			wantOk:        false,
		},
		{
			name:          "empty suffix",
			revision:      "v1.0.0+",
			wantCommitish: "",
			wantOk:        false,
		},
		{
			name:          "only a plus",
			revision:      "+",
			wantCommitish: "",
			wantOk:        false,
		},
		{
			name:          "second plus is part of the commitish check, not a separator",
			revision:      "v1.0.0+34ab+12cd",
			wantCommitish: "",
			wantOk:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commitish, ok := splitBuildMetadataRevision(tt.revision)
			if ok != tt.wantOk {
				t.Fatalf("splitBuildMetadataRevision(%q) ok = %v, want %v", tt.revision, ok, tt.wantOk)
			}
			if commitish != tt.wantCommitish {
				t.Fatalf("splitBuildMetadataRevision(%q) commitish = %q, want %q", tt.revision, commitish, tt.wantCommitish)
			}
		})
	}
}

func TestIsHex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "lowercase hex", input: "34ab12cd", want: true},
		{name: "uppercase hex", input: "34AB12CD", want: true},
		{name: "mixed case hex", input: "34Ab12Cd", want: true},
		{name: "digits only", input: "1234567890", want: true},
		{name: "empty string", input: "", want: true},
		{name: "contains non-hex letter", input: "34asdadasd", want: false},
		{name: "contains punctuation", input: "34ab-12cd", want: false},
		{name: "contains space", input: "34ab 12cd", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isHex(tt.input)
			if got != tt.want {
				t.Fatalf("isHex(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolveRevision_FullHash(t *testing.T) {
	fixture := newTestRepo(t, "v1.0.0", "v2.0.0")
	repo, err := (&ClonedRepo{Path: fixture.path}).openRepository()
	if err != nil {
		t.Fatalf("failed to open repository: %v", err)
	}

	hash, err := resolveRevision(repo, fixture.firstCommit)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if hash.String() != fixture.firstCommit {
		t.Fatalf("expected hash %q, got %q", fixture.firstCommit, hash.String())
	}
}

func TestResolveRevision_UnknownFullHash(t *testing.T) {
	fixture := newTestRepo(t, "v1.0.0", "v2.0.0")
	repo, err := (&ClonedRepo{Path: fixture.path}).openRepository()
	if err != nil {
		t.Fatalf("failed to open repository: %v", err)
	}

	unknown := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	_, err = resolveRevision(repo, unknown)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), unknown) {
		t.Fatalf("expected error to mention %q, got: %v", unknown, err)
	}
}

func TestResolveRevision_LightweightTag(t *testing.T) {
	fixture := newTestRepo(t, "v1.0.0", "v2.0.0")
	repo, err := (&ClonedRepo{Path: fixture.path}).openRepository()
	if err != nil {
		t.Fatalf("failed to open repository: %v", err)
	}

	hash, err := resolveRevision(repo, fixture.lightweightTag)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if hash.String() != fixture.firstCommit {
		t.Fatalf("expected hash %q, got %q", fixture.firstCommit, hash.String())
	}
}

func TestResolveRevision_AnnotatedTagPeelsToCommit(t *testing.T) {
	fixture := newTestRepo(t, "v1.0.0", "v2.0.0")
	repo, err := (&ClonedRepo{Path: fixture.path}).openRepository()
	if err != nil {
		t.Fatalf("failed to open repository: %v", err)
	}

	hash, err := resolveRevision(repo, fixture.annotatedTag)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if hash.String() != fixture.secondCommit {
		t.Fatalf("expected hash %q, got %q", fixture.secondCommit, hash.String())
	}
}

func TestResolveRevision_UnknownRef(t *testing.T) {
	fixture := newTestRepo(t, "v1.0.0", "v2.0.0")
	repo, err := (&ClonedRepo{Path: fixture.path}).openRepository()
	if err != nil {
		t.Fatalf("failed to open repository: %v", err)
	}

	_, err = resolveRevision(repo, "no-such-ref")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no-such-ref") {
		t.Fatalf("expected error to mention %q, got: %v", "no-such-ref", err)
	}
}

func TestResolveTargetHash_PrefersCommitishOverFullString(t *testing.T) {
	fixture := newTestRepo(t, "v1.0.0", "v2.0.0")
	c := &ClonedRepo{Path: fixture.path}
	repo, err := c.openRepository()
	if err != nil {
		t.Fatalf("failed to open repository: %v", err)
	}

	revision := fixture.annotatedTag + "+" + fixture.firstCommit[:8]
	hash, err := resolveTargetHash(repo, revision)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if hash.String() != fixture.firstCommit {
		t.Fatalf("expected hash %q, got %q", fixture.firstCommit, hash.String())
	}
}

func TestResolveTargetHash_FallsBackToFullStringWhenCommitishUnresolvable(t *testing.T) {
	fixture := newTestRepo(t, "v1.0.0", "v2.0.0")
	c := &ClonedRepo{Path: fixture.path}
	repo, err := c.openRepository()
	if err != nil {
		t.Fatalf("failed to open repository: %v", err)
	}

	revision := fixture.lightweightTag + "+deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	_, err = resolveTargetHash(repo, revision)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), revision) {
		t.Fatalf("expected error to contain full revision %q, got: %v", revision, err)
	}
}

func TestResolveTargetHash_FallsBackWhenSuffixIsNotHex(t *testing.T) {
	fixture := newTestRepo(t, "v1.0.0", "v2.0.0")
	c := &ClonedRepo{Path: fixture.path}
	repo, err := c.openRepository()
	if err != nil {
		t.Fatalf("failed to open repository: %v", err)
	}

	revision := fixture.lightweightTag + "+34asdadasd"
	_, err = resolveTargetHash(repo, revision)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), revision) {
		t.Fatalf("expected error to contain full revision %q, got: %v", revision, err)
	}
}
