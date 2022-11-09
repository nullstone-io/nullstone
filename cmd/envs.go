package cmd

import (
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/ryanuber/columnize"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/go-api-client.v0/ws"
	"gopkg.in/nullstone-io/nullstone.v0/runs"
	"math"
	"os"
	"sort"
	"strings"
	"time"
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
		EnvsDelete,
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
				envDetails[0] = "ID|Name|Type"
				for i, env := range envs {
					envDetails[i+1] = fmt.Sprintf("%d|%s|%s", env.Id, env.Name, strings.TrimSuffix(string(env.Type), "Env"))
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
		StackRequiredFlag,
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
			preview := c.IsSet("preview")

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

var EnvsDelete = &cli.Command{
	Name:        "delete",
	Description: "Deletes the given environment. Before issuing this command, make sure you have destroyed all infrastructure in the environment. If you are deleting a preview environment, you can use the `--force` flag to skip the confirmation prompt.",
	Usage:       "Create new environment",
	UsageText:   "nullstone envs delete --stack=<stack> --env=<env>	[--force]",
	Flags: []cli.Flag{
		StackRequiredFlag,
		EnvFlag,
		&cli.BoolFlag{
			Name:  "force",
			Usage: "Use this flag to skip the confirmation prompt when deleting an environment.",
		},
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			client := api.Client{Config: cfg}
			stackName := c.String("stack")
			envName := c.String("env")
			force := c.IsSet("force")

			stack, err := client.StacksByName().Get(stackName)
			if err != nil {
				return fmt.Errorf("error looking for stack %q: %w", stackName, err)
			} else if stack == nil {
				return fmt.Errorf("stack %q does not exist", stackName)
			}

			env, err := find.Env(cfg, stack.Id, envName)
			if err != nil {
				return fmt.Errorf("error looking for environment in stack %q - %q: %w", stack.Name, envName, err)
			} else if env == nil {
				return fmt.Errorf("environment %q does not exist in stack %q", envName, stack.Name)
			}

			if !force {
				fmt.Printf("You are about to delete an environment. Make sure you have destroyed all infrastructure in the environment before continuing.\n")
				confirm := []*survey.Question{
					{
						Name:     "Confirm",
						Validate: survey.Required,
						Prompt: &survey.Input{
							Message: "To confirm all infrastructure has been destroyed and you wish to continue, type 'delete':",
						},
					},
				}
				var confirmResponse string
				err := survey.Ask(confirm, &confirmResponse)
				if err != nil {
					return err
				}
				if confirmResponse != "delete" {
					fmt.Println("Deletion of the environment has been cancelled")
					return nil
				}
			}

			_, err = client.Environments().Destroy(stack.Id, env.Id)
			if err != nil {
				return fmt.Errorf("error deleting environment: %w", err)
			}

			fmt.Printf("Environment %s has been deleted\n", env.Name)
			return nil
		})
	},
}

func createPipelineEnv(client api.Client, stackId int64, name, providerName, region, zone string) error {
	if providerName == "" {
		return fmt.Errorf("provider is required")
	}
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
		StackRequiredFlag,
		EnvFlag,
		&cli.BoolFlag{
			Name:    "wait",
			Aliases: []string{"w"},
			Usage:   "Pass this flag in order to wait until the environment is provisioned. Progress will be logged in real time.",
		},
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			wait := c.IsSet("wait")
			return CancellableAction(func(ctx context.Context) error {
				if err := createEnvRun(ctx, c, cfg, false, wait); err != nil {
					return fmt.Errorf("error when trying to launch environment: %w", err)
				}
				return nil
			})
		})
	},
}

var EnvsDown = &cli.Command{
	Name:        "down",
	Description: "Destroys all the apps in an environment and all their dependent infrastructure. This command is useful for tearing down preview environments once you are finished with them.",
	Usage:       "Destroy an entire environment",
	UsageText:   "nullstone envs down --stack=<stack> --env=<env>",
	Flags: []cli.Flag{
		StackRequiredFlag,
		EnvFlag,
		&cli.BoolFlag{
			Name:    "wait",
			Aliases: []string{"w"},
			Usage:   "Pass this flag in order to wait until the environment is destroyed. Progress will be logged in real time.",
		},
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			wait := c.IsSet("wait")
			return CancellableAction(func(ctx context.Context) error {
				if err := createEnvRun(ctx, c, cfg, true, wait); err != nil {
					return fmt.Errorf("error when trying to destroy environment: %w", err)
				}
				return nil
			})
		})
	},
}

func createEnvRun(ctx context.Context, c *cli.Context, cfg api.Config, isDestroy, wait bool) error {
	client := api.Client{Config: cfg}
	stackName := c.String("stack")
	envName := c.String("env")

	action := "launch"
	if isDestroy {
		action = "destroy"
	}

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

	// body := types.CreateEnvRunInput{IsDestroy: isDestroy}
	// newRuns, err := client.EnvRuns().Create(stack.Id, env.Id, body)
	// if err != nil {
	// 	return fmt.Errorf("error creating run: %w", err)
	// }

	// if len(newRuns) <= 0 {
	// 	fmt.Fprintf(os.Stdout, "no runs created to %s the %q environment\n", action, envName)
	// 	return nil
	// }
	var newRuns []types.Run

	workspaces, err := client.Workspaces().List(stack.Id)
	if err != nil {
		return fmt.Errorf("error retrieving list of workspaces: %w", err)
	}
	blocks, err := client.Blocks().List(stack.Id)
	if err != nil {
		return fmt.Errorf("error retrieving list of blocks: %w", err)
	}

	for _, run := range newRuns {
		blockName := "(unknown)"
		workspace := findWorkspace(workspaces, run)
		if block := findBlock(blocks, workspace); block != nil {
			blockName = block.Name
		}
		browserUrl := ""
		if workspace != nil {
			browserUrl = fmt.Sprintf(" Logs: %s", runs.GetBrowserUrl(cfg, *workspace, run))
		}
		fmt.Fprintf(os.Stdout, "created run to %s %s and dependencies in %q environment.%s\n", action, blockName, envName, browserUrl)
	}

	return envStatus(ctx, cfg, stack, env, blocks, workspaces, wait)
}

func envStatus(ctx context.Context, cfg api.Config, stack *types.Stack, env *types.Environment, blocks []types.Block, workspaces []types.Workspace, wait bool) error {
	fmt.Fprintf(os.Stdout, "waiting for %q environment to be provisioned...\n", env.Name)

	client := api.Client{Config: cfg}
	msgs, err := client.Workspaces().Watch(ctx, stack.Id, ws.RetryInfinite(2*time.Second))
	if err != nil {
		return fmt.Errorf("error connecting to workspace updates stream: %w", err)
	}
	for msg := range msgs {
		if msg.Type == "error" {
			return fmt.Errorf(msg.Content)
		}
		if !wait && msg.Context == types.DeployPhaseWaitHealthy {
			// Stop streaming logs if we receive a log message from wait-healthy and no --wait
			break
		}
		fmt.Fprint(os.Stderr, msg.Content)
	}

	return nil
}

func findWorkspace(workspaces []types.Workspace, run types.Run) *types.Workspace {
	for _, workspace := range workspaces {
		if workspace.Uid == run.WorkspaceUid {
			return &workspace
		}
	}
	return nil
}
func findBlock(blocks []types.Block, workspace *types.Workspace) *types.Block {
	if workspace == nil {
		return nil
	}
	for _, block := range blocks {
		if workspace.BlockId == block.Id {
			return &block
		}
	}
	return nil
}
