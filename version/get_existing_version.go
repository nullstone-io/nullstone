package version

import (
	"context"
	"fmt"
	"slices"

	"github.com/nullstone-io/deployment-sdk/app"
	"gopkg.in/nullstone-io/nullstone.v0/vcs"
)

func GetExistingVersion(ctx context.Context, pusher app.Pusher, version string) (*Info, error) {
	result := Info{Version: version}
	var err error
	if result.CommitSha, err = vcs.GetCurrentCommitSha(); err != nil {
		return nil, fmt.Errorf("error calculating version: %w", err)
	}
	if result.Version == "" {
		result.Version = result.ShortCommitSha()
	}

	artifacts, err := pusher.ListArtifactVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("error retrieving list of artifact versions: %w", err)
	}
	if slices.Contains(artifacts, version) {
		return &result, nil
	}
	return nil, nil
}
