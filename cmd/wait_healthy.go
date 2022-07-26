package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"time"
)

func waitHealthy(ctx context.Context, cfg api.Config, appDetails app.Details, osWriters logging.OsWriters, provider *app.Provider, reference string) error {
	stdout := osWriters.Stdout()
	fmt.Fprintln(stdout, "Waiting for app to become healthy...")
	deployStatusGetter, err := provider.NewDeployStatusGetter(osWriters, cfg, appDetails)
	if err != nil {
		return fmt.Errorf("error creating app deployment status analyzer: %w", err)
	} else if deployStatusGetter == nil {
		// If we don't have a way of retrieving status, we are just going to complete the deployment
		return nil
	}
	if err := pollWaitHealthy(ctx, deployStatusGetter, reference); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "App is healthy")
	return nil
}

func pollWaitHealthy(ctx context.Context, deployStatusGetter app.DeployStatusGetter, reference string) error {
	for {
		status, err := deployStatusGetter.GetDeployStatus(ctx, reference)
		if err != nil {
			return fmt.Errorf("error querying app deployment status: %w", err)
		}
		switch status {
		case app.RolloutStatusComplete:
			return nil
		case app.RolloutStatusInProgress:
			return nil
		case app.RolloutStatusFailed:
			return fmt.Errorf("deployment failed")
		case app.RolloutStatusUnknown:
			return fmt.Errorf("unknown app deployment status")
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("cancelled")
		case <-time.After(5 * time.Second):
		}
	}
}
