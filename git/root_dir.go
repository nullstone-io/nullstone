package git

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"path/filepath"
)

func GetRootDir(curDir string) (string, *git.Repository, error) {
	if curDir == "" {
		curDir = "."
	}
	repo, err := git.PlainOpen(curDir)
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {
			return "", nil, nil
		}
		return "", nil, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		if errors.Is(err, git.ErrIsBareRepository) {
			return "", repo, nil
		}
		return "", repo, err
	}

	dir, err := filepath.Abs(worktree.Filesystem.Root())
	if err != nil {
		return "", nil, err
	}
	return dir, repo, nil
}
