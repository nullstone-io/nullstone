package deploy

import (
	"errors"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"log"
	"os"
)

var (
	ErrUnknownDeployer = errors.New("cannot perform deployment, unknown deployment pattern")
)

type Deployers []Deployer

func (d Deployers) Deploy(nsConfig api.Config, app *types.Application, workspace *types.Workspace, userConfig map[string]string) error {
	logger := log.New(os.Stderr, "", 0)

	if workspace.Status != types.WorkspaceStatusProvisioned {
		return fmt.Errorf("app %q has not been provisioned in %q environment yet", app.Name, workspace.EnvName)
	}
	if workspace.Module == nil {
		return fmt.Errorf("unknown module for workspace, cannot perform deployment")
	}

	for _, deployer := range d {
		if deployer.Detect(app, workspace) {
			logger.Printf("Identifying infrastructure for app %q\n", app.Name)
			infraConfig, err := deployer.Identify(nsConfig, app, workspace)
			if err != nil {
				return fmt.Errorf("Unable to identify app infrastructure: %w", err)
			}
			infraConfig.Print(logger)
			logger.Printf("Deploying app %q\n", app.Name)
			if err := deployer.Deploy(app, workspace, userConfig, infraConfig); err != nil {
				return fmt.Errorf("Unable to deploy app: %w", err)
			}
			logger.Printf("Deployed app %q\n", app.Name)
			return nil
		}
	}
	return ErrUnknownDeployer
}
