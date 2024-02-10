package vcs

func GetCurrentCommitSha() (string, error) {
	repo, err := GetGitRepo()
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
