package git

import (
	"github.com/go-git/go-git/v5"
	"net/url"
	"strings"
)

var (
	BlankVcsUrl = url.URL{
		Scheme: "https",
		Host:   "localhost",
		Path:   "local/repo",
	}
)

func GetVcsUrl(repo *git.Repository) url.URL {
	remote, err := repo.Remote("origin")
	if err != nil || remote == nil {
		return BlankVcsUrl
	}
	if remote.Config() == nil {
		return BlankVcsUrl
	}
	urls := remote.Config().URLs
	if len(urls) < 1 {
		return BlankVcsUrl
	}

	clean := strings.TrimSuffix(urls[0], ".git")
	if strings.HasPrefix(urls[0], "git@github.com:") {
		clean = strings.Replace(urls[0], "git@github.com:", "https://github.com/", 1)
	}

	u, err := url.Parse(clean)
	if u == nil {
		return BlankVcsUrl
	}
	return *u
}
