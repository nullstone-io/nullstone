package cmd

import (
	"context"
	"fmt"

	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/artifacts"
)

var (
	UniquePushFlag = &cli.BoolFlag{
		Name:  "unique",
		Usage: "Use this to *always* push the artifact with a unique version. If the input version already exists, an incrementing `-<count>` suffix is added.",
	}
)

// Push command performs a docker push to an authenticated image registry configured against an app/container
var Push = func(providers app.Providers) *cli.Command {
	return &cli.Command{
		Name:        "push",
		Description: "Upload (push) an artifact containing the source for your application. Specify a semver version to associate with the artifact. The version specified can be used in the deploy command to select this artifact. By default, this command does nothing if an artifact with the same version already exists. Use --unique to force push with a unique version.",
		Usage:       "Push artifact",
		UsageText:   "nullstone push [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			OldEnvFlag,
			AppSourceFlag,
			AppVersionFlag,
			UniquePushFlag,
		},
		Action: func(c *cli.Context) error {
			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				osWriters := CliOsWriters{Context: c}
				source := c.String(AppSourceFlag.Name)

				pusher, err := getPusher(providers, cfg, appDetails)
				if err != nil {
					return err
				}

				info, err := calcPushInfo(ctx, c, pusher)
				if err != nil {
					return err
				}

				if err := recordArtifact(ctx, osWriters, cfg, appDetails, info); err != nil {
					return err
				}

				return push(ctx, osWriters, pusher, source, info)
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

func calcPushInfo(ctx context.Context, c *cli.Context, pusher app.Pusher) (artifacts.VersionInfo, error) {
	osWriters := CliOsWriters{Context: c}
	stderr := osWriters.Stderr()
	version, unique := c.String(AppVersionFlag.Name), c.IsSet(UniquePushFlag.Name)

	if version == "" {
		fmt.Fprintf(stderr, "No version specified. Defaulting version based on current git commit sha...\n")
	}
	info, err := artifacts.GetVersionInfoFromWorkingDir(version)
	if err != nil {
		return info, err
	}
	if version == "" {
		fmt.Fprintf(stderr, "Version defaulted to %q\n", info.DesiredVersion)
	}

	deconflictor, err := artifacts.NewVersionDeconflictor(ctx, pusher)
	if err != nil {
		return info, fmt.Errorf("error reading artifact registry: %w", err)
	}

	if unique {
		info.EffectiveVersion = deconflictor.CreateUnique(info.DesiredVersion)
		fmt.Fprintf(stderr, "Artifacts matching %q exist in artifact registry. Changing version to %q.\n", info.DesiredVersion, info.EffectiveVersion)
	} else if deconflictor.DoesVersionExist(info.DesiredVersion) {
		fmt.Fprintln(stderr, "App artifact already exists. Skipped push.")
		fmt.Fprintln(stderr, "")
		return info, nil
	} else {
		info.EffectiveVersion = info.DesiredVersion
	}
	return info, nil
}

func recordArtifact(ctx context.Context, osWriters logging.OsWriters, cfg api.Config, appDetails app.Details, info artifacts.VersionInfo) error {
	apiClient := api.Client{Config: cfg}
	if _, err := apiClient.CodeArtifacts().Upsert(ctx, appDetails.App.StackId, appDetails.App.Id, appDetails.Env.Id, info.EffectiveVersion, info.CommitInfo); err != nil {
		fmt.Fprintf(osWriters.Stderr(), "Unable to record artifact in Nullstone: %s\n", err)
	}
	fmt.Fprintf(osWriters.Stderr(), "Recorded artifact (%s) in Nullstone (commit SHA = %s).\n", info.EffectiveVersion, info.CommitInfo.CommitSha)
	return nil
}

func push(ctx context.Context, osWriters logging.OsWriters, pusher app.Pusher, source string, info artifacts.VersionInfo) error {
	fmt.Fprintln(osWriters.Stderr(), "Pushing app artifact...")
	if err := pusher.Push(ctx, source, info.EffectiveVersion); err != nil {
		return fmt.Errorf("error pushing artifact: %w", err)
	}
	fmt.Fprintln(osWriters.Stderr(), "App artifact pushed.")
	fmt.Fprintln(osWriters.Stderr(), "")
	return nil
}
