package cmd

import (
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func CreateDeploy(nsConfig api.Config, appDetails app.Details, commitSha, version string) (*types.Deploy, error) {
	client := api.Client{Config: nsConfig}
	payload := api.DeployCreatePayload{
		FromSource: false,
		Version:    version,
		CommitSha:  commitSha,
	}
	newDeploy, err := client.Deploys().Create(appDetails.App.StackId, appDetails.App.Id, appDetails.Env.Id, payload)
	if err != nil {
		return nil, fmt.Errorf("error creating deploy: %w", err)
	} else if newDeploy == nil {
		return nil, fmt.Errorf("unable to create deploy")
	}
	return newDeploy, nil
}
