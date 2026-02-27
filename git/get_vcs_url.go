package git

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
)

func GetVcsUrl(repo *git.Repository) (*url.URL, error) {
	if repo == nil {
		return nil, nil
	}

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

	clean, err := ParseRemote(urls[0])
	if err != nil {
		return nil, fmt.Errorf("error parsing remote url %q: %w", urls[0], err)
	}
	clean.Path = strings.TrimSuffix(clean.Path, ".git")
	return clean, nil
}

var scpLike = regexp.MustCompile(`^([^@]+@)?([^:]+):(.+)$`)

// ParseRemote parses a remote URL into a url.URL
// This converts standard URL formats and SCP-style formats into a URL format
// This does not resolve host aliases (e.g. github-brad-sickles -> github.com)
func ParseRemote(raw string) (*url.URL, error) {
	// --- Case 1: Standard URL ---
	if strings.Contains(raw, "://") {
		return url.Parse(raw)
	}

	// --- Case 2: SCP-style ---
	if m := scpLike.FindStringSubmatch(raw); m != nil {
		user := strings.TrimSuffix(m[1], "@")
		return &url.URL{
			Scheme: "git",
			User:   url.User(user),
			Host:   m[2],
			Path:   m[3],
		}, nil
	}

	return nil, fmt.Errorf("remote url format is not supported")
}
