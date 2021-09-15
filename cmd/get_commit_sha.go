package cmd

import "gopkg.in/nullstone-io/nullstone.v0/vcs"

func getCurrentCommitSha() (string, error) {
	repo, err := vcs.GetGitRepo()
	if err != nil {
		return "", err
	}
	if repo == nil {
		return "", nil
	}
	ref, err := repo.Head()
	if err != nil {
		return "", err
	}
	hash := ref.Hash()
	if hash.IsZero() {
		return "", nil
	}
	return hash.String(), nil
}
