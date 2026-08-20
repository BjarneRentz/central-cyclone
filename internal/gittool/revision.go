package gittool

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
)

const (
	minCommitishHexLen = 4
	maxCommitishHexLen = 40
)

func resolveTargetHash(repo *git.Repository, revision string) (plumbing.Hash, error) {
	if commitish, ok := splitBuildMetadataRevision(revision); ok {
		if hash, err := resolveRevision(repo, commitish); err == nil {
			return hash, nil
		}
	}

	return resolveRevision(repo, revision)
}

func resolveRevision(repo *git.Repository, target string) (plumbing.Hash, error) {
	if plumbing.IsHash(target) {
		hash := plumbing.NewHash(target)
		if _, err := repo.CommitObject(hash); err != nil {
			return plumbing.ZeroHash, fmt.Errorf("failed to resolve revision %q: %w", target, err)
		}
		return hash, nil
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(target))
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to resolve revision %q: %w", target, err)
	}
	targetHash := *hash

	tagObj, err := repo.TagObject(targetHash)
	if err == nil {
		commit, err := tagObj.Commit()
		if err != nil {
			return plumbing.ZeroHash, fmt.Errorf("failed to get commit from annotated tag: %w", err)
		}
		targetHash = commit.Hash
	}

	return targetHash, nil
}

func splitBuildMetadataRevision(revision string) (commitish string, ok bool) {
	idx := strings.IndexByte(revision, '+')
	if idx == -1 {
		return "", false
	}

	commitish = revision[idx+1:]
	if len(commitish) < minCommitishHexLen || len(commitish) > maxCommitishHexLen || !isHex(commitish) {
		return "", false
	}

	return commitish, true
}

func isHex(s string) bool {
	for _, r := range s {
		isDigit := r >= '0' && r <= '9'
		isLower := r >= 'a' && r <= 'f'
		isUpper := r >= 'A' && r <= 'F'
		if !isDigit && !isLower && !isUpper {
			return false
		}
	}
	return true
}
