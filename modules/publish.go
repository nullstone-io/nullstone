package modules

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/mitchellh/colorstring"
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
func Publish(ctx context.Context, cfg api.Config, logger *log.Logger, input PublishInput) (*PublishOutput, error) {
	logger.Println(fmt.Sprintf("Reading module manifest file %q", manifestFilename))
	manifest, err := ManifestFromFile(manifestFilename)
	if err != nil {
		return nil, err
	}
	logger.Println()

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
		timestamp := time.Now().Unix()
		version = fmt.Sprintf("%s+%s.%d", version, commitSha, timestamp)
	}

	version = strings.TrimPrefix(version, "v")
	if isValid := semver.IsValid(fmt.Sprintf("v%s", version)); !isValid {
		return nil, fmt.Errorf("version %q is not a valid semver", version)
	}

	// Package module files into tar.gz
	allIncludes := append(input.Includes, manifest.Includes...)
	tarballFilename, err := Package(logger, manifest, version, allIncludes)
	if err != nil {
		return nil, err
	}
	logger.Println()

	logger.Println("Publishing module...")
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
	colorstring.Fprintln(logger.Writer(), fmt.Sprintf("[green]Published %s/%s@%s", manifest.OrgName, manifest.Name, version))
	logger.Println()

	return &PublishOutput{
		Version:  version,
		Manifest: manifest,
	}, nil
}
