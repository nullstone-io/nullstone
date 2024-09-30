package version

type Info struct {
	CommitSha string
	Version   string
}

func (i Info) ShortCommitSha() string {
	maxLength := 7
	if len(i.CommitSha) < maxLength {
		maxLength = len(i.CommitSha)
	}
	return i.CommitSha[0:maxLength]
}
