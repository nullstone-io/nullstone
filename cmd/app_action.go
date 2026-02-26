package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// AppOrAppWorkspace finds the application and optionally the environment based on CLI flags.
// If --env is not specified, only the application is returned (env will be nil).
// If --env is specified, both the application and its environment are returned.
func AppOrAppWorkspace(c *cli.Context, cfg api.Config) (*types.Application, *types.Environment, error) {
	ctx := context.TODO()
	client := api.Client{Config: cfg}

	stackName := c.String(StackRequiredFlag.Name)
	appName := c.String(RequiredAppFlag.Name)
	envName := c.String(EnvOptionalFlag.Name)

	app, err := find.App(ctx, cfg, appName, stackName)
	if err != nil {
		return nil, nil, fmt.Errorf("error looking for app %q: %w", appName, err)
	} else if app == nil {
		return nil, nil, fmt.Errorf("app %q does not exist in stack %q", appName, stackName)
	}

	if envName == "" {
		return app, nil, nil
	}

	stack, err := client.StacksByName().Get(ctx, stackName)
	if err != nil {
		return nil, nil, fmt.Errorf("error looking for stack %q: %w", stackName, err)
	} else if stack == nil {
		return nil, nil, fmt.Errorf("stack %q does not exist", stackName)
	}

	env, err := find.Env(ctx, cfg, stack.Id, envName)
	if err != nil {
		return nil, nil, fmt.Errorf("error looking for environment %q: %w", envName, err)
	} else if env == nil {
		return nil, nil, fmt.Errorf("environment %q does not exist in stack %q", envName, stackName)
	}

	return app, env, nil
}
