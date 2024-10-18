package git

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"path/filepath"
)

func GetRootDir(curDir string) (string, error) {
	if curDir == "" {
		curDir = "."
	}
	repo, err := git.PlainOpen(curDir)
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {
			return "", nil
		}
		return "", err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		if errors.Is(err, git.ErrIsBareRepository) {
			return "", nil
		}
		return "", err
	}

	return filepath.Abs(worktree.Filesystem.Root())
}
