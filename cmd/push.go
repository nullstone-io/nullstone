package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/vcs"
	version2 "gopkg.in/nullstone-io/nullstone.v0/version"
)

// Push command performs a docker push to an authenticated image registry configured against an app/container
var Push = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:        "push",
		Description: "Upload (push) an artifact containing the source for your application. Specify a semver version to associate with the artifact. The version specified can be used in the deploy command to select this artifact.",
		Usage:       "Push artifact",
		UsageText:   "nullstone push [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			OldEnvFlag,
			AppSourceFlag,
			AppVersionFlag,
		},
		Action: func(c *cli.Context) error {
			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				osWriters := CliOsWriters{Context: c}
				source, version := c.String("source"), c.String("version")

				pusher, err := getPusher(providers, cfg, appDetails)
				if err != nil {
					return err
				}

				if version == "" {
					fmt.Fprintf(osWriters.Stderr(), "No version specified. Defaulting version based on current git commit sha...\n")
					_, version, err = calcNewVersion(ctx, pusher)
					if err != nil {
						return err
					}
					fmt.Fprintf(osWriters.Stderr(), "Version defaulted to: %s\n", version)
				}

				return push(ctx, osWriters, pusher, source, version)
			})
		},
	}
}

func getPusher(providers app.Providers, cfg api.Config, appDetails app.Details) (app.Pusher, error) {
	ctx := context.TODO()
	pusher, err := providers.FindPusher(ctx, logging.StandardOsWriters{}, outputs.ApiRetrieverSource{Config: cfg}, appDetails)
	if err != nil {
		return nil, fmt.Errorf("error creating app pusher: %w", err)
	} else if pusher == nil {
		return nil, fmt.Errorf("this application category=%s, type=%s does not support push", appDetails.Module.Category, appDetails.Module.Type)
	}
	return pusher, nil
}

func calcNewVersion(ctx context.Context, pusher app.Pusher) (string, string, error) {
	shortSha, err := vcs.GetCurrentShortCommitSha()
	if err != nil {
		return "", "", fmt.Errorf("error calculating version: %w", err)
	}

	artifacts, err := pusher.ListArtifactVersions(ctx)
	if err != nil {
		// if we aren't able to pull the list of artifact versions, we can just use the short sha as the fallback
		return shortSha, shortSha, nil
	}

	seq := version2.FindLatestVersionSequence(shortSha, artifacts)
	if err != nil {
		return shortSha, "", fmt.Errorf("error calculating version: %w", err)
	}

	version := shortSha
	// -1 means we didn't find any existing deploys for this commitSha
	// and we will just use the shortSha as the version
	// otherwise we will append the sequence number
	if seq > -1 {
		version = fmt.Sprintf("%s-%d", shortSha, seq+1)
	}

	return shortSha, version, nil
}

func push(ctx context.Context, osWriters logging.OsWriters, pusher app.Pusher, source, version string) error {
	fmt.Fprintln(osWriters.Stderr(), "Pushing app artifact...")
	if err := pusher.Push(ctx, source, version); err != nil {
		return fmt.Errorf("error pushing artifact: %w", err)
	}
	fmt.Fprintln(osWriters.Stderr(), "App artifact pushed.")
	fmt.Fprintln(osWriters.Stderr(), "")
	return nil
}
