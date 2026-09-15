package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

var IacSyncRepoFlag = &cli.StringFlag{
	Name:  "repo",
	Usage: "Source repository that owns the synced blocks, as <owner>/<name> or a repository URL (e.g. --repo=nullstone-io/nullstone). Defaults to the origin remote of the current git repository.",
}

// errMissingSyncRepo is returned when neither --repo nor the git origin remote identifies
// the repository the sync runs for. The server rejects such a sync, and it could never update
// blocks another sync already owns, so it is not worth submitting.
var errMissingSyncRepo = errors.New("could not determine the source repository from git (not a git clone, no origin remote, or an unsupported remote url); pass --repo=<owner/name>")

// resolveIacSyncRepo picks the repository an IaC sync runs for: an explicit --repo wins,
// otherwise the origin remote the commit info was read from. Both produce the canonical
// https://<host>/<owner>/<name> form the server records on synced apps and events.
func resolveIacSyncRepo(flagValue string, ci types.CommitInfo) (types.Repo, error) {
	if flagValue = strings.TrimSpace(flagValue); flagValue != "" {
		repo, err := types.RepoFromUrl(flagValue)
		if err != nil {
			return types.Repo{}, fmt.Errorf("invalid --repo: %w", err)
		}
		repo.Name = strings.TrimSuffix(repo.Name, ".git")
		if repo.Owner == "" || repo.Name == "" {
			return types.Repo{}, fmt.Errorf("invalid --repo %q: must be <owner>/<name>", flagValue)
		}
		repo.Url = fmt.Sprintf("https://%s/%s/%s", repo.Host, repo.Owner, repo.Name)
		repo.InferVcsProvider()
		return repo, nil
	}
	if ci.Repository.Owner != "" && ci.Repository.Name != "" {
		return ci.Repository, nil
	}
	return types.Repo{}, errMissingSyncRepo
}
