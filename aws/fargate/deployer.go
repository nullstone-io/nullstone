package fargate

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/deploy"
	"os"
)

var _ deploy.Deployer = Deployer{}

// Deployer will deploy an app/container using the following workflow to a fargate cluster:
//   Get task definition
//   Change image tag in task definition
//   Register new task definition
//   Deregister old task definition
//   Update ECS Service
type Deployer struct{}

func (d Deployer) Detect(app *types.Application, workspace *types.Workspace) bool {
	if workspace.Module.Category != types.CategoryAppContainer {
		return false
	}
	if workspace.Module.Type != "service/aws-fargate" {
		return false
	}
	return true
}

func (d Deployer) Identify(nsConfig api.Config, app *types.Application, workspace *types.Workspace) (deploy.InfraConfig, error) {
	return newInfraConfig(nsConfig, workspace)
}

func (d Deployer) Deploy(app *types.Application, workspace *types.Workspace, config map[string]string, infraConfig interface{}) error {
	ic := infraConfig.(*InfraConfig)

	taskDef, err := ic.GetTaskDefinition()
	if err != nil {
		return fmt.Errorf("error retrieving current service information: %w", err)
	}

	taskDefArn := *taskDef.TaskDefinitionArn
	if imageTag := config["imageTag"]; imageTag != "" {
		fmt.Fprintf(os.Stderr, "Updating image tag to %q\n", imageTag)
		newTaskDef, err := ic.UpdateTaskImageTag(taskDef, imageTag)
		if err != nil {
			return fmt.Errorf("error updating task with new image tag: %w", err)
		}
		taskDefArn = *newTaskDef.TaskDefinitionArn
	}

	if err := ic.UpdateServiceTask(taskDefArn); err != nil {
		return fmt.Errorf("error deploying service: %w", err)
	}
	return nil
}
