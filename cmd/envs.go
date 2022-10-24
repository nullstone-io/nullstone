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
	Description: "Creates a new environment in the given stack. The environment will be created as the last environment in the pipeline. Specify the provider, region, and zone to determine where infrastructure will be provisioned for this environment.",
	Usage:       "Create new environment",
	UsageText:   "nullstone envs new --name=<name> --stack=<stack> --provider=<provider>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "name",
			Usage:    "Provide a name for this new environment",
			Required: true,
		},
		StackFlag,
		&cli.StringFlag{
			Name:     "provider",
			Usage:    "Provide the name of the provider to use for this environment",
			Required: true,
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

			stack, err := client.StacksByName().Get(stackName)
			if err != nil {
				return fmt.Errorf("error looking for stack %q: %w", stackName, err)
			} else if stack == nil {
				return fmt.Errorf("stack %q does not exist", stackName)
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
					Region:       c.String("region"),
				}
				if pc.Aws.Region == "" {
					pc.Aws.Region = awsDefaultRegion
				}
			case "gcp":
				pc.Gcp = &types.GcpProviderConfig{
					ProviderName: provider.Name,
					Region:       c.String("region"),
					Zone:         c.String("zone"),
				}
				if pc.Gcp.Region == "" || pc.Gcp.Zone == "" {
					pc.Gcp.Region = gcpDefaultRegion
					pc.Gcp.Zone = gcpDefaultZone
				}
			default:
				return fmt.Errorf("CLI does not support provider type %q yet", provider.ProviderType)
			}

			env, err := client.Environments().Create(stack.Id, &types.Environment{
				Name:           name,
				ProviderConfig: pc,
			})
			if err != nil {
				return fmt.Errorf("error creating stack: %w", err)
			}
			fmt.Printf("created %q environment\n", env.Name)
			return nil
		})
	},
}
