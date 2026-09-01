package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
)

var RunEnvVarFlag = &cli.StringSliceFlag{
	Name:    "env-var",
	Aliases: []string{"e"},
	Usage: `Pass environment variables to the job/task. 
You can use this flag multiple times to specify multiple environment variables. 
For each environment variable, use <NAME>=<value>.

This is supported for AWS ECS/Fargate jobs, GCP Cloud Run jobs, and Kubernetes jobs.
This is not supported for AWS Lambda; a lambda function's environment variables are configured at deploy time.
Use --payload to send input to a lambda function.
`,
	Required: false,
}

var RunPayloadFlag = &cli.StringFlag{
	Name: "payload",
	Usage: `Pass an input event to the function.
Use '@filename' to read the payload from a file, or '-' to read the payload from stdin.
The payload must be a valid JSON document.

This is supported for AWS Lambda functions.
`,
	Required: false,
}

var Run = func(appProviders app.Providers, providers admin.Providers) *cli.Command {
	return &cli.Command{
		Name:        "run",
		Description: "Starts a new container/serverless for the given Nullstone job/task. ",
		Usage:       "Starts a new job/task",
		UsageText:   "nullstone run [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options] [-- command [args...]]",
		Flags: []cli.Flag{
			StackFlag,
			AppFlag,
			EnvFlag,
			ContainerFlag,
			RunEnvVarFlag,
			RunPayloadFlag,
		},
		Action: func(c *cli.Context) error {
			var cmd []string
			if c.Args().Present() {
				cmd = c.Args().Slice()
			}

			payload, err := readRunPayload(c.String(RunPayloadFlag.Name))
			if err != nil {
				return err
			}

			return AppWorkspaceAction(c, func(ctx context.Context, cfg api.Config, appDetails app.Details) error {
				username, err := getCurrentUsername(ctx, cfg)
				if err != nil {
					return err
				}

				rawEnvVars := c.StringSlice(RunEnvVarFlag.Name)
				envVars := make(map[string]string)
				for _, raw := range rawEnvVars {
					before, after, ok := strings.Cut(raw, "=")
					if !ok {
						return fmt.Errorf("invalid --env-var flag, expected <NAME>=<value>")
					}
					envVars[before] = after
				}
				envVars[admin.TriggerEnvVar] = admin.TriggerManual
				envVars[admin.TriggerNameEnvVar] = username

				source := outputs.ApiRetrieverSource{Config: cfg}
				osWriters := logging.StandardOsWriters{}

				logStreamer, err := appProviders.FindLogStreamer(ctx, osWriters, source, appDetails)
				if err != nil {
					return err
				}

				remoter, err := providers.FindRemoter(ctx, osWriters, source, appDetails)
				if err != nil {
					return err
				} else if remoter == nil {
					module := appDetails.Module
					platform := strings.TrimSuffix(fmt.Sprintf("%s:%s", module.Platform, module.Subplatform), ":")
					return fmt.Errorf("The Nullstone CLI does not currently support the `run` command for the %q application. (Module = %s/%s, App Category = app/%s, Platform = %s)",
						appDetails.App.Name, module.OrgName, module.Name, module.Subcategory, platform)
				}
				options := admin.RunOptions{
					Container:   c.String(ContainerFlag.Name),
					Payload:     payload,
					Username:    username,
					LogStreamer: logStreamer,
					LogEmitter:  app.NewWriterLogEmitter(os.Stdout),
				}
				return remoter.Run(ctx, options, cmd, envVars)
			})
		},
	}
}

// readRunPayload resolves the --payload flag
// The payload is read from a file when prefixed with '@' and from stdin when set to '-'
func readRunPayload(raw string) ([]byte, error) {
	switch {
	case raw == "":
		return nil, nil
	case raw == "-":
		payload, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("unable to read the payload from stdin: %w", err)
		}
		return payload, nil
	case strings.HasPrefix(raw, "@"):
		filename := raw[1:]
		payload, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("unable to read the payload from file (%s): %w", filename, err)
		}
		return payload, nil
	default:
		return []byte(raw), nil
	}
}
