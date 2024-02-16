package vcs

func GetCurrentShortCommitSha() (string, error) {
	sha, err := GetCurrentCommitSha()
	if err != nil {
		return "", err
	}
	maxLength := 7
	if len(sha) < maxLength {
		maxLength = len(sha)
	}
	return sha[0:maxLength], nil
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
