package version

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"gopkg.in/nullstone-io/nullstone.v0/vcs"
)

func GetCurrent(ctx context.Context, pusher app.Pusher) (Info, error) {
	result := Info{}
	var err error
	if result.CommitSha, err = vcs.GetCurrentCommitSha(); err != nil {
		return Info{}, fmt.Errorf("error calculating version: %w", err)
	}

	artifacts, err := pusher.ListArtifactVersions(ctx)
	if err != nil {
		// if we aren't able to pull the list of artifact versions, we can just use the short sha as the fallback
		result.Version = result.ShortCommitSha()
		return result, nil
	}

	seq := FindLatestVersionSequence(result.ShortCommitSha(), artifacts)
	if err != nil {
		result.Version = ""
		return result, fmt.Errorf("error calculating version: %w", err)
	}

	// no existing deploys found for this commitSha
	if seq == -1 {
		result.Version = ""
		return result, fmt.Errorf("no artifacts found for this commit SHA (%s) - you must perform a successful push before deploying", result.ShortCommitSha())
	}
	// only one deploy found for this commitSha, so we don't need to append a sequence
	if seq == 0 {
		result.Version = result.ShortCommitSha()
		return result, nil
	}
	
	result.Version = fmt.Sprintf("%s-%d", result.ShortCommitSha(), seq)
	return result, nil
}
