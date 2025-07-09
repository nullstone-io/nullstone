package vcs

import (
	"errors"
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
	if errors.Is(err, git.ErrRepositoryNotExists) {
		return nil, nil
	}
	return repo, err
}
