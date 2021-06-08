package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"os"
	"os/signal"
	"syscall"
)

type AppDetails struct {
	App       *types.Application
	Env       *types.Environment
	Workspace *types.Workspace
}

type AppActionFn func(ctx context.Context, cfg api.Config, provider app.Provider, details AppDetails) error

func AppAction(c *cli.Context, providers app.Providers, fn AppActionFn) error {
	_, cfg, err := SetupProfileCmd(c)
	if err != nil {
		return err
	}

	if c.NArg() != 2 {
		cli.ShowCommandHelp(c, "logs")
		return fmt.Errorf("invalid usage")
	}
	appName := c.Args().Get(0)
	envName := c.Args().Get(1)

	finder := NsFinder{Config: cfg}
	app, env, workspace, err := finder.GetAppAndWorkspace(appName, c.String("stack-name"), envName)
	if err != nil {
		return err
	}

	provider := providers.Find(workspace.Module.Category, workspace.Module.Type)
	if provider == nil {
		return fmt.Errorf("unable to push, this CLI does not support category=%s, type=%s", workspace.Module.Category, workspace.Module.Type)
	}

	ctx := context.Background()
	// Handle Ctrl+C, kill stream
	ctx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-term
		cancelFn()
	}()

	return fn(ctx, cfg, provider, AppDetails{
		App:       app,
		Env:       env,
		Workspace: workspace,
	})
}
