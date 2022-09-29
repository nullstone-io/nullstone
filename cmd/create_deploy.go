package cmd

import (
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func CreateDeploy(nsConfig api.Config, appDetails app.Details, version string) (*types.Deploy, error) {
	client := api.Client{Config: nsConfig}
	newDeploy, err := client.Deploys().Create(appDetails.App.StackId, appDetails.App.Id, appDetails.Env.Id, version, false)
	if err != nil {
		return nil, fmt.Errorf("error creating deploy: %w", err)
	} else if newDeploy == nil {
		return nil, fmt.Errorf("unable to create deploy")
	}
	return newDeploy, nil
}
