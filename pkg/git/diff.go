package git

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func Diff(repoDir, base, head string) (string, error) {
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return "", fmt.Errorf("git open: %w", err)
	}
	baseCommit, err := repo.CommitObject(plumbing.NewHash(base))
	if err != nil {
		return "", fmt.Errorf("git base commit: %w", err)
	}
	baseTree, err := baseCommit.Tree()
	if err != nil {
		return "", fmt.Errorf("git base tree: %w", err)
	}
	headCommit, err := repo.CommitObject(plumbing.NewHash(head))
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

func PrefixDiff(diff string) string {
	var prefixed strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(diff))
	for i := 0; scanner.Scan(); i++ {
		line := scanner.Text()
		prefixed.WriteString(fmt.Sprintf("%d: %s\n", i, line))
	}
	return prefixed.String()
}
