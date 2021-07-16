package cmd

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type AppEnvActionFn func(ctx context.Context, cfg api.Config, provider app.Provider, details app.Details) error

func AppEnvAction(c *cli.Context, providers app.Providers, fn AppEnvActionFn) error {
	_, cfg, err := SetupProfileCmd(c)
	if err != nil {
		return err
	}

	if c.NArg() != 2 {
		cli.ShowCommandHelp(c, c.Command.Name)
		return fmt.Errorf("invalid usage")
	}
	appName := c.Args().Get(0)
	envName := c.Args().Get(1)
	stackName := c.String("stack-name")

	logger := log.New(os.Stderr, "", 0)
	logger.Printf("Performing application command (Org=%s, App=%s, Stack=%s, Env=%s)", cfg.OrgName, appName, stackName, envName)
	logger.Println()

	finder := NsFinder{Config: cfg}
	application, env, workspace, err := finder.GetAppAndWorkspace(appName, stackName, envName)
	if err != nil {
		return err
	}

	provider := providers.Find(workspace.Module.Category, workspace.Module.Type)
	if provider == nil {
		return fmt.Errorf("this CLI does not support application category=%s, type=%s", workspace.Module.Category, workspace.Module.Type)
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

	return fn(ctx, cfg, provider, app.Details{
		App:       application,
		Env:       env,
		Workspace: workspace,
	})
}
