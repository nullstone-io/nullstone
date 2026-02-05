package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

type AppWorkspaceFn func(ctx context.Context, cfg api.Config, appDetails app.Details) error

func AppWorkspaceAction(c *cli.Context, fn AppWorkspaceFn) error {
	_, cfg, err := SetupProfileCmd(c)
	if err != nil {
		return err
	}

	return ParseAppEnv(c, true, func(stackName, appName, envName string) error {
		logger := log.New(c.App.ErrWriter, "", 0)
		logger.Printf("Performing application command (Org=%s, App=%s, Stack=%s, Env=%s)", cfg.OrgName, appName, stackName, envName)
		logger.Println()

		ctx := context.TODO()
		apiClient := api.Client{Config: cfg}
		infraDetails, err := apiClient.WorkspaceInfraDetails().GetByName(ctx, stackName, appName, envName, false)
		if err != nil {
			return err
		}

		application, ok := infraDetails.Block().(types.Application)
		if !ok {
			logger.Printf("This command operates on Applications, but the Block is a(n) %s", infraDetails.BlockType())
			return fmt.Errorf("cannot run command")
		}

		return CancellableAction(func(ctx context.Context) error {
			return fn(ctx, cfg, app.Details{
				App:       &application,
				Env:       &infraDetails.Env,
				Workspace: &infraDetails.Workspace,
				Module:    &infraDetails.Module,
			})
		})
	})
}
