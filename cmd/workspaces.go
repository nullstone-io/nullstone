package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/nullstone.v0/tfconfig"
	"gopkg.in/nullstone-io/nullstone.v0/workspaces"
	"path"
)

var (
	backendFilename         = "__backend__.tf"
	activeWorkspaceFilename = path.Join(".nullstone", "active-workspace.yml")
)

var Workspaces = &cli.Command{
	Name:      "workspaces",
	Usage:     "View and modify workspaces",
	UsageText: "nullstone workspaces [subcommand]",
	Subcommands: []*cli.Command{
		WorkspacesSelect,
	},
}

var WorkspacesSelect = &cli.Command{
	Name:      "select",
	Usage:     "Select workspace",
	UsageText: "nullstone workspaces select [--stack=<stack>] --block=<block> --env=<env>",
	Flags: []cli.Flag{
		StackFlag,
		&cli.StringFlag{
			Name:     "block",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "env",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		return ProfileAction(c, func(cfg api.Config) error {
			if !tfconfig.IsCredsConfigured(cfg) {
				if err := tfconfig.ConfigCreds(cfg); err != nil {
					fmt.Printf("Warning: unable to configure Terraform-based credentials with Nullstone servers: %s\n", err)
				} else {
					fmt.Println("Configured Terraform-based credentials with Nullstone servers.")
				}
			}

			client := api.Client{Config: cfg}
			stackName := c.String("stack")
			blockName := c.String("block")
			envName := c.String("env")
			sbe, err := find.StackBlockEnvByName(cfg, stackName, blockName, envName)
			if err != nil {
				return err
			}

			// TODO: Add support for capability testing --> workspaces.Manifest.CapabilityId

			targetWorkspace := workspaces.Manifest{
				OrgName:   cfg.OrgName,
				StackId:   sbe.Stack.Id,
				StackName: sbe.Stack.Name,
				BlockId:   sbe.Block.Id,
				BlockName: sbe.Block.Name,
				BlockRef:  sbe.Block.Reference,
				EnvId:     sbe.Env.Id,
				EnvName:   sbe.Env.Name,
			}
			workspace, err := client.Workspaces().Get(targetWorkspace.StackId, targetWorkspace.BlockId, targetWorkspace.EnvId)
			if err != nil {
				return err
			} else if workspace == nil {
				return fmt.Errorf("could not find workspace (stack=%s, block=%s, env=%s)", stackName, blockName, envName)
			}
			targetWorkspace.WorkspaceUid = workspace.Uid.String()

			if err := workspaces.WriteBackendTf(cfg, workspace.Uid, backendFilename); err != nil {
				return fmt.Errorf("error writing terraform backend file: %w", err)
			}
			if err := targetWorkspace.WriteToFile(activeWorkspaceFilename); err != nil {
				return fmt.Errorf("error writing active workspace file: %w", err)
			}

			fmt.Printf(`Selected workspace:
  Stack:     %s
  Block:     %s
  Env:       %s
  Workspace: %s
`, sbe.Stack.Name, sbe.Block.Name, sbe.Env.Name, workspace.Uid)

			return CancellableAction(func(ctx context.Context) error {
				if err := workspaces.Init(ctx); err != nil {
					fallbackMessage := `Unable to initialize terraform.
Reset .terraform/ directory and run 'terraform init'.`
					fmt.Println(fallbackMessage)
				}
				return nil
			})
		})
	},
}
