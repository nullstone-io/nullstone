package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func TestResolveIacSyncRepo(t *testing.T) {
	fromGit := types.CommitInfo{Repository: types.Repo{
		Provider: "github", Host: "github.com", Owner: "acme", Name: "from-git", Url: "https://github.com/acme/from-git",
	}}
	github := func(owner, name string) types.Repo {
		return types.Repo{Provider: "github", Host: "github.com", Owner: owner, Name: name, Url: "https://github.com/" + owner + "/" + name}
	}

	tests := []struct {
		name    string
		flag    string
		ci      types.CommitInfo
		want    types.Repo
		wantErr string
	}{
		{name: "git origin when no flag", ci: fromGit, want: fromGit.Repository},
		{name: "flag owner/name", flag: "acme/repo", ci: fromGit, want: github("acme", "repo")},
		{name: "flag host/owner/name", flag: "gitlab.com/acme/repo",
			want: types.Repo{Provider: "gitlab", Host: "gitlab.com", Owner: "acme", Name: "repo", Url: "https://gitlab.com/acme/repo"}},
		{name: "flag https url", flag: "https://github.com/acme/repo", want: github("acme", "repo")},
		{name: "flag https url with .git", flag: "https://github.com/acme/repo.git", want: github("acme", "repo")},
		{name: "flag is trimmed", flag: "  acme/repo  ", want: github("acme", "repo")},
		{name: "flag with a single token", flag: "repo", wantErr: "invalid --repo"},
		{name: "flag with too many tokens", flag: "https://github.com/acme/repo/extra", wantErr: "invalid --repo"},
		{name: "nothing to go on", wantErr: "pass --repo=<owner/name>"},
		{name: "git gave only an owner", ci: types.CommitInfo{Repository: types.Repo{Owner: "acme"}}, wantErr: "pass --repo=<owner/name>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveIacSyncRepo(tt.flag, tt.ci)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Equal(t, types.Repo{}, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
