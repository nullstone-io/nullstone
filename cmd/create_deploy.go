package cmd

import (
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

func CreateDeploy(nsConfig api.Config, appDetails app.Details, version string) error {
	client := api.Client{Config: nsConfig}
	result, err := client.Deploys().Create(appDetails.App.StackId, appDetails.App.Id, appDetails.Env.Id, version)
	if err != nil {
		return fmt.Errorf("error updating app version: %w", err)
	} else if result == nil {
		return fmt.Errorf("could not find application environment")
	}
	return nil
}
