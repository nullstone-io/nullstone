package fargate

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"log"
)

func DeployContainer(app *types.Application, workspace *types.Workspace) error {
	if workspace.LastSuccessfulRun == nil || workspace.LastSuccessfulRun.Apply == nil {
		return fmt.Errorf("cannot find outputs for application")
	}
	apply := workspace.LastSuccessfulRun.Apply
	log.Println(apply.Outputs)
	return nil

	// Get task definition
	// Change image tag in task definition
	// Register new task definition
	// Deregister old task definition
	// Update ECS Service
}
