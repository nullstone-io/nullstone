package version

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"gopkg.in/nullstone-io/nullstone.v0/vcs"
)

func CalcNew(ctx context.Context, pusher app.Pusher) (Info, error) {
	result := Info{}
	var err error
	if result.CommitSha, err = vcs.GetCurrentCommitSha(); err != nil {
		return Info{}, fmt.Errorf("error calculating version: %w", err)
	}
	result.Version = result.ShortCommitSha()

	artifacts, err := pusher.ListArtifactVersions(ctx)
	if err != nil {
		// if we aren't able to pull the list of artifact versions, we can just use the short sha as the fallback
		return result, nil
	}

	seq := FindLatestVersionSequence(result.ShortCommitSha(), artifacts)
	if err != nil {
		result.Version = ""
		return result, fmt.Errorf("error calculating version: %w", err)
	}

	// -1 means we didn't find any existing deploys for this commitSha
	// and we will just use the shortSha as the version
	// otherwise we will append the sequence number
	if seq > -1 {
		result.Version = fmt.Sprintf("%s-%d", result.ShortCommitSha(), seq+1)
	}

	return result, nil
}
