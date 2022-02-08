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

	appName := c.String(AppFlag.Name)
	envName := c.String(EnvFlag.Name)

	// TODO: `nullstone cmd <app> <env>` is deprecated
	// Drop the following block to parse this format once fully removed
	// This format is only parsed if --app and --env are empty
	if appName == "" && envName == "" {
		if c.NArg() >= 2 {
			appName = c.Args().Get(0)
			envName = c.Args().Get(1)
		}
	}

	if appName == "" || envName == "" {
		cli.ShowCommandHelp(c, c.Command.Name)
		return fmt.Errorf("'--app' and '--env' flags are required to run this command")
	}

	stackName := c.String(StackFlag.Name)
	specifiedStack := stackName
	if specifiedStack == "" {
		specifiedStack = "<unspecified>"
	}

	logger := log.New(os.Stderr, "", 0)
	logger.Printf("Performing application command (Org=%s, App=%s, Stack=%s, Env=%s)", cfg.OrgName, appName, specifiedStack, envName)
	logger.Println()

	finder := NsFinder{Config: cfg}
	appDetails, err := finder.FindAppDetails(appName, stackName, envName)
	if err != nil {
		return err
	}

	provider := providers.Find(appDetails.Module.Category, appDetails.Module.Type)
	if provider == nil {
		return fmt.Errorf("this CLI does not support application category=%s, type=%s", appDetails.Module.Category, appDetails.Module.Type)
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

	return fn(ctx, cfg, provider, appDetails)
}
