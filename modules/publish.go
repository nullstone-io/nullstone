package modules

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"golang.org/x/mod/semver"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/vcs"
)

var manifestFilename = path.Join(".nullstone", "module.yml")

type PublishInput struct {
	// Version is the semver version to publish.
	// Special values: "next-patch", "next-build".
	// Empty string defaults to a build version "0.0.0-<short-sha>".
	Version string
	// Includes specifies additional file patterns to package (beyond manifest.Includes).
	Includes []string
}

type PublishOutput struct {
	Version  string
	Manifest *types.ModuleManifest
}

// Publish packages and publishes a module from the current working directory.
// It reads the manifest from .nullstone/module.yml, resolves the version,
// packages the module files, and uploads to the registry.
func Publish(ctx context.Context, cfg api.Config, input PublishInput) (*PublishOutput, error) {
	manifest, err := ManifestFromFile(manifestFilename)
	if err != nil {
		return nil, err
	}

	version := input.Version

	if version == "next-patch" {
		version, err = NextPatch(ctx, cfg, manifest)
		if err != nil {
			return nil, err
		}
	}

	if version == "next-build" {
		version, err = NextPatch(ctx, cfg, manifest)
		if err != nil {
			return nil, err
		}
		commitSha, err := vcs.GetCurrentShortCommitSha()
		if err != nil || commitSha == "" {
			return nil, fmt.Errorf("using next-build requires a git repository with a commit: %w", err)
		}
		version = fmt.Sprintf("%s+%s", version, commitSha)
	}

	version = strings.TrimPrefix(version, "v")
	if isValid := semver.IsValid(fmt.Sprintf("v%s", version)); !isValid {
		return nil, fmt.Errorf("version %q is not a valid semver", version)
	}

	// Package module files into tar.gz
	allIncludes := append(input.Includes, manifest.Includes...)
	tarballFilename, err := Package(manifest, version, allIncludes)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "Created module package %q\n", tarballFilename)

	// Open tarball to publish
	tarball, err := os.Open(tarballFilename)
	if err != nil {
		return nil, err
	}
	defer tarball.Close()

	client := api.Client{Config: cfg}
	if err := client.ModuleVersions().Create(ctx, manifest.OrgName, manifest.Name, manifest.ToolName, version, tarball); err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "Published %s/%s@%s\n", manifest.OrgName, manifest.Name, version)

	return &PublishOutput{
		Version:  version,
		Manifest: manifest,
	}, nil
}
