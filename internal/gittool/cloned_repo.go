package gittool

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
)

type ClonedRepo struct {
	Path    string
	RepoUrl string
	repo    *git.Repository
}

// openRepository opens the git repository at the cloned path
func (c *ClonedRepo) openRepository() (*git.Repository, error) {
	if c.repo != nil {
		return c.repo, nil
	}

	repo, err := git.PlainOpen(c.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	c.repo = repo
	return repo, nil
}

func (c *ClonedRepo) Pull() error {
	repo, err := c.openRepository()
	if err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	slog.Info("📥 Updating repository", "repo", c.RepoUrl)

	token := os.Getenv("GIT_TOKEN")
	pullOpts := &git.PullOptions{}

	if token != "" {
		pullOpts.Auth = &http.BasicAuth{
			Username: "git",
			Password: token,
		}
	}

	err = w.Pull(pullOpts)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull repository: %w", err)
	}

	return nil
}

func (c *ClonedRepo) CheckoutRevision(revision string) error {
	repo, err := c.openRepository()
	if err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	slog.Info("🔄 Preparing checkout", "repo", c.RepoUrl, "revision", revision)

	var targetHash plumbing.Hash

	// 1. FIRST CHOICE: If it's a direct 40-character SHA-1 hash, use it immediately.
	// This completely bypasses the "reference not found" lookup error.
	if plumbing.IsHash(revision) {
		targetHash = plumbing.NewHash(revision)
	} else {
		// 2. SECOND CHOICE: Resolve it as a named reference (Tag, Branch, Short-Hash)
		hash, err := repo.ResolveRevision(plumbing.Revision(revision))
		if err != nil {
			return fmt.Errorf("failed to resolve revision %q: %w", revision, err)
		}
		targetHash = *hash

		// 3. FIX FOR ANNOTATED TAGS: Peel the tag to get the real commit hash
		tagObj, err := repo.TagObject(targetHash)
		if err == nil {
			commit, err := tagObj.Commit()
			if err != nil {
				return fmt.Errorf("failed to get commit from annotated tag: %w", err)
			}
			targetHash = commit.Hash
		}
	}

	slog.Info("🏷️  Checking out", "repo", c.RepoUrl, "hash", targetHash.String())

	// Perform the checkout using Force to ensure a clean slate
	err = w.Checkout(&git.CheckoutOptions{
		Hash:  targetHash,
		Force: true,
	})
	if err != nil {
		return fmt.Errorf("failed to checkout revision: %w", err)
	}

	return nil
}

// GetCurrentRevision returns the current commit hash
func (c *ClonedRepo) GetCurrentRevision() (string, error) {
	repo, err := c.openRepository()
	if err != nil {
		return "", err
	}

	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	return ref.Hash().String(), nil
}

// Tries to update the repo and returns true, if an update was available
func (c *ClonedRepo) UpdateIfAvailable() (bool, error) {
	repo, err := c.openRepository()
	if err != nil {
		return false, err
	}

	w, err := repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	slog.Info("📥 Updating repository", "repo", c.RepoUrl)

	token := os.Getenv("GIT_TOKEN")
	pullOpts := &git.PullOptions{}

	if token != "" {
		pullOpts.Auth = &http.BasicAuth{
			Username: "git",
			Password: token,
		}
	}
	err = w.Pull(pullOpts)

	if err != nil && err == git.NoErrAlreadyUpToDate {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil

}
