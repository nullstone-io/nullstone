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
	skipWhenExistsFlag := &cli.BoolFlag{
		Name:  "skip-when-exists",
		Usage: "Skip pushing if the artifact already exists in the registry",
	}

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
			skipWhenExistsFlag,
		},
		Action: func(c *cli.Context) error {
			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				osWriters := CliOsWriters{Context: c}
				source, version := c.String("source"), c.String("version")
				skipWhenExists := c.IsSet(skipWhenExistsFlag.Name)

				pusher, err := getPusher(providers, cfg, appDetails)
				if err != nil {
					return err
				}

				if skipWhenExists {
					fmt.Fprintf(osWriters.Stderr(), "Checking to see if artifact version already exists...\n")
					info, err := version2.GetExistingVersion(ctx, pusher, version)
					if err != nil {
						return fmt.Errorf("error checking if version already exists: %w", err)
					} else if info != nil {
						fmt.Fprintln(osWriters.Stderr(), "App artifact already exists. Skipped push.")
						fmt.Fprintln(osWriters.Stderr(), "")
						return nil
					}
					version = info.Version
				} else if version == "" {
					fmt.Fprintf(osWriters.Stderr(), "No version specified. Defaulting version based on current git commit sha...\n")
					info, err := version2.CalcNew(ctx, pusher)
					if err != nil {
						return err
					}
					version = info.Version
					fmt.Fprintf(osWriters.Stderr(), "Version defaulted to: %s\n", version)
				}

				if err := recordArtifact(ctx, osWriters, cfg, appDetails, version); err != nil {
					return err
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

func recordArtifact(ctx context.Context, osWriters logging.OsWriters, cfg api.Config, appDetails app.Details, version string) error {
	commitInfo, err := vcs.GetCommitInfo()
	if err != nil {
		return fmt.Errorf("error retrieving commit info from .git/: %w", err)
	}
	apiClient := api.Client{Config: cfg}
	if _, err := apiClient.CodeArtifacts().Upsert(ctx, appDetails.App.StackId, appDetails.App.Id, appDetails.Env.Id, version, commitInfo); err != nil {
		fmt.Fprintf(osWriters.Stderr(), "Unable to record artifact in Nullstone: %s\n", err)
	}
	fmt.Fprintf(osWriters.Stderr(), "Recorded artifact (%s) in Nullstone (commit SHA = %s).\n", version, commitInfo.CommitSha)
	return nil
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
