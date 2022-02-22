package git

import "github.com/go-git/go-git/v5"

func RepoFromDir(dir string) *git.Repository {
	repo, err := git.PlainOpen(".")
	if err == git.ErrRepositoryNotExists {
		return nil
	}
	return repo
}
