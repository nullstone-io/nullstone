package git

import (
	"github.com/go-git/go-git/v5"
	"net/url"
	"strings"
)

func GetVcsUrl(repo *git.Repository) (*url.URL, error) {
	remote, err := repo.Remote("origin")
	if err != nil {
		return nil, err
	}
	if remote == nil || remote.Config() == nil {
		return nil, nil
	}
	urls := remote.Config().URLs
	if len(urls) < 1 {
		return nil, nil
	}

	clean := strings.TrimSuffix(urls[0], ".git")
	if strings.HasPrefix(urls[0], "git@github.com:") {
		clean = strings.Replace(urls[0], "git@github.com:", "https://github.com/", 1)
	}

	return url.Parse(clean)
}
