package modules

import (
	"context"
	"fmt"
	"golang.org/x/mod/semver"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"strconv"
	"strings"
)

// NextPatch bumps the patch in major.minor.patch from the latest module version
func NextPatch(ctx context.Context, cfg api.Config, manifest *types.ModuleManifest) (string, error) {
	latestVersion, err := find.ModuleVersion(ctx, cfg, fmt.Sprintf("%s/%s", manifest.OrgName, manifest.Name), "latest")
	if err != nil {
		return "", fmt.Errorf("error retrieving latest version: %w", err)
	} else if latestVersion == nil {
		return BumpPatch("v0.0.0"), nil
	}
	return BumpPatch(latestVersion.Version), nil
}

func BumpPatch(version string) string {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	curPatch := getPatch(version)
	return fmt.Sprintf("%s.%d", semver.MajorMinor(version), curPatch+1)
}

func getPatch(version string) int {
	tokens := strings.SplitN(semver.Canonical(version), ".", 3)
	if len(tokens) < 3 {
		return 0
	}
	rawPatch := strings.TrimSuffix(tokens[2], semver.Prerelease(version))
	patch, err := strconv.Atoi(rawPatch)
	if err != nil {
		return 0
	}
	return patch
}
