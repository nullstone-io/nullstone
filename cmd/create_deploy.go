package cmd

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"gopkg.in/nullstone-io/go-api-client.v0"
)

func CreateDeploy(nsConfig api.Config, appDetails app.Details, commitSha, version string) (*api.DeployCreateResult, error) {
	ctx := context.TODO()
	client := api.Client{Config: nsConfig}
	payload := api.DeployCreatePayload{
		FromSource: false,
		Version:    version,
		CommitSha:  commitSha,
	}
	result, err := client.Deploys().Create(ctx, appDetails.App.StackId, appDetails.App.Id, appDetails.Env.Id, payload)
	if err != nil {
		return nil, fmt.Errorf("error creating deploy: %w", err)
	} else if result == nil {
		return nil, fmt.Errorf("unable to create deploy")
	}
	return result, nil
}
