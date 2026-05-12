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
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/vcs"
)

// ConditionChecksumChanged is the only currently-supported value for
// PublishInput.Condition. When set, Publish fetches the latest published version
// and skips uploading if its checksum matches the freshly-packaged tarball.
const ConditionChecksumChanged = "checksum-changed"

var manifestFilename = path.Join(".nullstone", "module.yml")

type PublishInput struct {
	// Version is the semver version to publish.
	// Special values: "next-patch", "next-build".
	// Empty string defaults to a build version "0.0.0-<short-sha>".
	Version string
	// Includes specifies additional file patterns to package (beyond manifest.Includes).
	Includes []string
	// Condition gates whether the publish actually uploads. Currently only
	// ConditionChecksumChanged is supported; empty string disables the gate.
	Condition string
}

type PublishOutput struct {
	Version  string
	Manifest *types.ModuleManifest
	// Skipped is true when Condition caused Publish to short-circuit before
	// uploading. Version reflects the existing latest version in that case.
	Skipped bool
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
	logger.SetPrefix("    ")
	WriteManifestToLogger(*manifest, logger)
	logger.SetPrefix("")
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
	tarballFilename, checksum, err := Package(logger, manifest, version, allIncludes)
	if err != nil {
		return nil, err
	}
	logger.Println()

	if input.Condition == ConditionChecksumChanged {
		moduleSource := fmt.Sprintf("%s/%s", manifest.OrgName, manifest.Name)
		latest, err := find.ModuleVersion(ctx, cfg, moduleSource, "latest")
		if err != nil {
			return nil, fmt.Errorf("error looking up latest version for %s: %w", moduleSource, err)
		}
		switch {
		case latest == nil:
			logger.Println("no existing versions found, publishing initial version")
		case latest.Checksum == "":
			logger.Println(fmt.Sprintf("latest version v%s has no checksum recorded, publishing to backfill", latest.Version))
		case latest.Checksum == checksum:
			colorstring.Fprintln(logger.Writer(), fmt.Sprintf("[yellow]checksum matches latest published version v%s — skipping publish", latest.Version))
			return &PublishOutput{
				Version:  latest.Version,
				Manifest: manifest,
				Skipped:  true,
			}, nil
		}
	}

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
