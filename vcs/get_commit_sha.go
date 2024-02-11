package vcs

func GetCurrentShortCommitSha() (string, error) {
	sha, err := GetCurrentCommitSha()
	if err != nil {
		return "", err
	}
	if len(sha) < 7 {
		return "", nil
	}
	return sha[0:7], nil
}

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
