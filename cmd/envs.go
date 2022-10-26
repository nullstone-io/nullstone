package cmd

import (
	"fmt"
	"github.com/ryanuber/columnize"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"math"
	"sort"
)

var (
	awsDefaultRegion = "us-east-1"
	gcpDefaultRegion = "us-east1"
	gcpDefaultZone   = "us-east1b"
)

var Envs = &cli.Command{
	Name:      "envs",
	Usage:     "View and modify environments",
	UsageText: "nullstone envs [subcommand]",
	Subcommands: []*cli.Command{
		EnvsList,
		EnvsNew,
		EnvsUp,
		EnvsDown,
	},
}

var EnvsList = &cli.Command{
	Name:        "list",
	Description: "Shows a list of the environments for the given stack. Set the `--detail` flag to show more details about each environment.",
	Usage:       "List environments",
	UsageText:   "nullstone envs list --stack=<stack-name>",
	Flags: []cli.Flag{
		StackRequiredFlag,
		&cli.BoolFlag{
			Name:    "detail",
			Aliases: []string{"d"},
			Usage:   "Use this flag to show more details about each environment",
		},
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			stackName := c.String(StackRequiredFlag.Name)
			stack, err := find.Stack(cfg, stackName)
			if err != nil {
				return fmt.Errorf("error retrieving stack: %w", err)
			} else if stack == nil {
				return fmt.Errorf("stack %s does not exist", stackName)
			}

			client := api.Client{Config: cfg}
			envs, err := client.Environments().List(stack.Id)
			if err != nil {
				return fmt.Errorf("error listing environments: %w", err)
			}
			sort.SliceStable(envs, func(i, j int) bool {
				var first int
				if envs[i].PipelineOrder == nil {
					first = math.MaxInt
				} else {
					first = *envs[i].PipelineOrder
				}
				var second int
				if envs[j].PipelineOrder == nil {
					second = math.MaxInt
				} else {
					second = *envs[j].PipelineOrder
				}
				return first < second
			})

			if c.IsSet("detail") {
				envDetails := make([]string, len(envs)+1)
				envDetails[0] = "ID|Name"
				for i, env := range envs {
					envDetails[i+1] = fmt.Sprintf("%d|%s", env.Id, env.Name)
				}
				fmt.Println(columnize.Format(envDetails, columnize.DefaultConfig()))
			} else {
				for _, env := range envs {
					fmt.Println(env.Name)
				}
			}

			return nil
		})
	},
}

var EnvsNew = &cli.Command{
	Name:        "new",
	Description: "Creates a new environment in the given stack. If the `--preview` parameter is set, a preview environment will be created and the `--provider` parameter will not be used. Otherwise, a standard environment will be created as the last environment in the pipeline. Specify the provider, region, and zone to determine where infrastructure will be provisioned for this environment.",
	Usage:       "Create new environment",
	UsageText:   "nullstone envs new --name=<name> --stack=<stack> [--provider=<provider>] [--preview]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "name",
			Usage:    "Provide a name for this new environment",
			Required: true,
		},
		StackFlag,
		&cli.BoolFlag{
			Name:  "preview",
			Usage: "Use this flag to create a preview environment. If not set, a standard environment will be created.",
		},
		&cli.StringFlag{
			Name:  "provider",
			Usage: "Select the name of the provider to use for this environment. When creating a preview environment, this parameter will not be used.",
		},
		&cli.StringFlag{
			Name:  "region",
			Usage: fmt.Sprintf("Select which region to launch infrastructure for this environment. Defaults to %s for AWS and %s for GCP.", awsDefaultRegion, gcpDefaultRegion),
		},
		&cli.StringFlag{
			Name:  "zone",
			Usage: fmt.Sprintf("For GCP, select the zone to launch infrastructure for this environment. Defaults to %s", gcpDefaultZone),
		},
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}
			name := c.String("name")
			stackName := c.String("stack")
			providerName := c.String("provider")
			region := c.String("region")
			zone := c.String("zone")
			preview := c.Bool("preview")

			stack, err := client.StacksByName().Get(stackName)
			if err != nil {
				return fmt.Errorf("error looking for stack %q: %w", stackName, err)
			} else if stack == nil {
				return fmt.Errorf("stack %q does not exist", stackName)
			}

			if preview {
				return createPreviewEnv(client, stack.Id, name)
			} else {
				return createPipelineEnv(client, stack.Id, name, providerName, region, zone)
			}
		})
	},
}

func createPipelineEnv(client api.Client, stackId int64, name, providerName, region, zone string) error {
	provider, err := client.Providers().Get(providerName)
	if err != nil {
		return fmt.Errorf("error looking for provider %q: %w", providerName, err)
	} else if provider == nil {
		return fmt.Errorf("provider %q does not exist", providerName)
	}

	pc := types.ProviderConfig{}
	switch provider.ProviderType {
	case "aws":
		pc.Aws = &types.AwsProviderConfig{
			ProviderName: provider.Name,
			Region:       region,
		}
		if pc.Aws.Region == "" {
			pc.Aws.Region = awsDefaultRegion
		}
	case "gcp":
		pc.Gcp = &types.GcpProviderConfig{
			ProviderName: provider.Name,
			Region:       region,
			Zone:         zone,
		}
		if pc.Gcp.Region == "" || pc.Gcp.Zone == "" {
			pc.Gcp.Region = gcpDefaultRegion
			pc.Gcp.Zone = gcpDefaultZone
		}
	default:
		return fmt.Errorf("CLI does not support provider type %q yet", provider.ProviderType)
	}

	env, err := client.Environments().Create(stackId, &types.Environment{
		Name:           name,
		ProviderConfig: pc,
	})
	if err != nil {
		return fmt.Errorf("error creating environment: %w", err)
	}
	fmt.Printf("created %q environment\n", env.Name)
	return nil
}

func createPreviewEnv(client api.Client, stackId int64, name string) error {
	env, err := client.PreviewEnvs().Create(stackId, &api.CreatePreviewEnvInput{Name: name})
	if err != nil {
		return fmt.Errorf("error creating preview environment: %w", err)
	}
	fmt.Printf("created %q preview environment\n", env.Name)
	return nil
}

var EnvsUp = &cli.Command{
	Name:        "up",
	Description: "Launches an entire environment including all of its apps. This command can be used to stand up an entire preview environment.",
	Usage:       "Launch an entire environment",
	UsageText:   "nullstone envs up --stack=<stack> --env=<env>",
	Flags: []cli.Flag{
		StackFlag,
		EnvFlag,
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			return createEnvRun(c, cfg, false)
		})
	},
}

var EnvsDown = &cli.Command{
	Name:        "down",
	Description: "Destroys all the apps in an environment and all their dependent infrastructure. This command is useful for tearing down preview environments once you are finished with them.",
	Usage:       "Destroy an entire environment",
	UsageText:   "nullstone envs down --stack=<stack> --env=<env>",
	Flags: []cli.Flag{
		StackFlag,
		EnvFlag,
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			return createEnvRun(c, cfg, true)
		})
	},
}

func createEnvRun(c *cli.Context, cfg api.Config, isDestroy bool) error {
	client := api.Client{Config: cfg}
	stackName := c.String("stack")
	envName := c.String("env")

	stack, err := client.StacksByName().Get(stackName)
	if err != nil {
		return fmt.Errorf("error looking for stack %q: %w", stackName, err)
	} else if stack == nil {
		return fmt.Errorf("stack %q does not exist", stackName)
	}

	env, err := find.Env(cfg, stack.Id, envName)
	if err != nil {
		return fmt.Errorf("error looking for environment in stack %d - %q: %w", stack.Id, envName, err)
	} else if env == nil {
		return fmt.Errorf("environment %q does not exist in stack %d", envName, stack.Id)
	}

	body := types.CreateEnvRunInput{IsDestroy: isDestroy}
	runs, err := client.EnvRuns().Create(stack.Id, env.Id, body)
	if err != nil {
		if isDestroy {
			return fmt.Errorf("error creating run to destroy environment: %w", err)
		} else {
			return fmt.Errorf("error creating run to launch environment: %w", err)
		}
	}
	if isDestroy {
		fmt.Printf("created run to destroy %d apps in the %q environment\n", len(runs), envName)
	} else {
		fmt.Printf("created run to launch %d apps in the %q environment\n", len(runs), envName)
	}
	return nil
}
