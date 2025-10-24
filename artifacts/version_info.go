package artifacts

import (
	"context"
	"fmt"
	"slices"

	"github.com/nullstone-io/deployment-sdk/app"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/vcs"
)

type VersionInfo struct {
	DesiredVersion   string
	EffectiveVersion string
	CommitInfo       types.CommitInfo
}

func GetVersionInfoFromWorkingDir(desiredVersion string) (VersionInfo, error) {
	info := VersionInfo{DesiredVersion: desiredVersion}

	var err error
	if info.CommitInfo, err = vcs.GetCommitInfo(); err != nil {
		return VersionInfo{}, fmt.Errorf("error retrieving commit info from .git/: %w", err)
	}
	if info.DesiredVersion == "" {
		info.DesiredVersion = shortCommitSha(info.CommitInfo.CommitSha)
	}

	return info, nil
}

type VersionDeconflictor struct {
	versions []string
}

func NewVersionDeconflictor(ctx context.Context, pusher app.Pusher) (*VersionDeconflictor, error) {
	versions, err := pusher.ListArtifactVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("error retrieving list of artifact versions: %w", err)
	}
	return &VersionDeconflictor{versions: versions}, nil
}

// CreateUnique calculates a new version from the input version if the version already exists
// If the version already exists, the version is returned unchanged
func (d *VersionDeconflictor) CreateUnique(version string) string {
	seq := FindLatestVersionSequence(version, d.versions)

	// -1 means we didn't find any existing versions matching the input version
	// use the input version
	if seq == -1 {
		return version
	}

	// Otherwise, we need to calculate a new version
	return fmt.Sprintf("%s-%d", version, seq+1)
}

func (d *VersionDeconflictor) DoesVersionExist(version string) bool {
	return slices.Contains(d.versions, version)
}

func shortCommitSha(commitSha string) string {
	maxLength := 7
	if len(commitSha) < maxLength {
		maxLength = len(commitSha)
	}
	return commitSha[0:maxLength]
}
