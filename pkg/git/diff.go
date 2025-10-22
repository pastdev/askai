package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func Diff(repoDir, base, head string) (string, error) {
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return "", fmt.Errorf("git open: %w", err)
	}
	baseRev, err := repo.ResolveRevision(plumbing.Revision(base))
	if err != nil {
		return "", fmt.Errorf("git base rev: %w", err)
	}
	baseCommit, err := repo.CommitObject(*baseRev)
	if err != nil {
		return "", fmt.Errorf("git base commit: %w", err)
	}
	baseTree, err := baseCommit.Tree()
	if err != nil {
		return "", fmt.Errorf("git base tree: %w", err)
	}
	headRev, err := repo.ResolveRevision(plumbing.Revision(head))
	if err != nil {
		return "", fmt.Errorf("git head rev: %w", err)
	}
	headCommit, err := repo.CommitObject(*headRev)
	if err != nil {
		return "", fmt.Errorf("git head commit: %w", err)
	}
	headTree, err := headCommit.Tree()
	if err != nil {
		return "", fmt.Errorf("git head tree: %w", err)
	}
	patch, err := baseTree.Patch(headTree)
	if err != nil {
		return "", fmt.Errorf("git diff: %w", err)
	}
	return patch.String(), nil
}
