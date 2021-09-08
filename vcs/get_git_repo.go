package vcs

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"os"
)

func GetGitRepo() (*git.Repository, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("unable to find git repository: %w", err)
	}
	repo, err := git.PlainOpenWithOptions(curDir, &git.PlainOpenOptions{
		DetectDotGit:          true,
		EnableDotGitCommonDir: true,
	})
	if err == git.ErrRepositoryNotExists {
		return nil, nil
	}
	return repo, err
}
