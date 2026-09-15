package vcs

import (
	"testing"

	"github.com/go-git/go-git/v5/config"
	"github.com/stretchr/testify/assert"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func TestParseRemoteRepository(t *testing.T) {
	// stands in for ssh -G: only the alias nullstone uses resolves
	resolve := func(alias string) string {
		if alias == "github-nullstone" {
			return "github.com"
		}
		return ""
	}
	github := func(owner, name string) types.Repo {
		return types.Repo{Provider: "github", Host: "github.com", Owner: owner, Name: name, Url: "https://github.com/" + owner + "/" + name}
	}

	tests := []struct {
		name string
		raw  string
		want types.Repo
	}{
		{"scp-style with user", "git@github.com:acme/repo.git", github("acme", "repo")},
		{"scp-style without .git", "git@github.com:acme/repo", github("acme", "repo")},
		// NUL-209: the remote that produced a sync with no repository
		{"scp-style host alias without user", "github-nullstone:astraeus-wealth-tech/devops.git", github("astraeus-wealth-tech", "devops")},
		{"scp-style host alias with user", "git@github-nullstone:acme/repo.git", github("acme", "repo")},
		{"ssh url", "ssh://git@github.com/acme/repo.git", github("acme", "repo")},
		{"ssh url with port", "ssh://git@github.com:22/acme/repo.git", github("acme", "repo")},
		{"https", "https://github.com/acme/repo.git", github("acme", "repo")},
		{"https without .git", "https://github.com/acme/repo", github("acme", "repo")},
		{"https with token user", "https://x-access-token:abc@github.com/acme/repo.git", github("acme", "repo")},
		{"http", "http://github.com/acme/repo.git", github("acme", "repo")},
		{"git protocol", "git://github.com/acme/repo.git", github("acme", "repo")},
		{"trailing whitespace from git remote get-url", "git@github.com:acme/repo.git\n", github("acme", "repo")},
		{"gitlab subgroup keeps the full owner path", "git@gitlab.com:acme/platform/repo.git",
			types.Repo{Provider: "gitlab", Host: "gitlab.com", Owner: "acme/platform", Name: "repo", Url: "https://gitlab.com/acme/platform/repo"}},
		{"unknown alias stays as-is rather than guessing a provider", "git@work:acme/repo.git",
			types.Repo{Host: "work", Owner: "acme", Name: "repo", Url: "https://work/acme/repo"}},
		// nothing usable: the zero value, never a partial repo with a bogus url
		{"empty", "", types.Repo{}},
		{"scp-style with no path", "git@github.com", types.Repo{}},
		{"https with only an owner", "https://github.com/acme", types.Repo{}},
		{"https with no path", "https://github.com", types.Repo{}},
		{"local path", "/srv/git/repo.git", types.Repo{}},
		{"file url", "file:///srv/git/acme/repo.git", types.Repo{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				assert.Equal(t, tt.want, parseRemoteRepository(tt.raw, resolve))
			})
		})
	}
}

func TestExtractApiRepository(t *testing.T) {
	assert.Equal(t, types.Repo{}, extractApiRepository(nil))
	assert.Equal(t, types.Repo{}, extractApiRepository(&config.RemoteConfig{Name: "origin"}))
	got := extractApiRepository(&config.RemoteConfig{Name: "origin", URLs: []string{"https://github.com/acme/repo.git"}})
	assert.Equal(t, "acme", got.Owner)
	assert.Equal(t, "repo", got.Name)
	assert.Equal(t, "https://github.com/acme/repo", got.Url)
}
