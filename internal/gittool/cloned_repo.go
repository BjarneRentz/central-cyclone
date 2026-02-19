package gittool

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v6"
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

// CheckoutTag checks out a specific tag
func (c *ClonedRepo) CheckoutTag(tag string) error {
	repo, err := c.openRepository()
	if err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	slog.Info("🏷️  Checking out tag", "repo", c.RepoUrl, "tag", tag)

	// Resolve tag to its hash
	tagRef, err := repo.Tag(tag)
	if err != nil {
		return fmt.Errorf("failed to find tag %q: %w", tag, err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash: tagRef.Hash(),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout tag: %w", err)
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
