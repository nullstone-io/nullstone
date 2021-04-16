package aws_fargate

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"log"
	"os"
)

var _ app.Provider = Provider{}

type Provider struct {
}

func (p Provider) Push(nsConfig api.Config, app *types.Application, workspace *types.Workspace, userConfig map[string]string) error {
	panic("implement me")
}

// Deploy takes the following steps to deploy an AWS Fargate service
//   Get task definition
//   Change image tag in task definition
//   Register new task definition
//   Deregister old task definition
//   Update ECS Service (This always causes deployment)
func (p Provider) Deploy(nsConfig api.Config, app *types.Application, workspace *types.Workspace, userConfig map[string]string) error {
	logger := log.New(os.Stderr, "", 0)

	logger.Printf("Identifying infrastructure for app %q\n", app.Name)
	ic, err := discoverInfraConfig(nsConfig, workspace)
	if err != nil {
		return fmt.Errorf("Unable to identify app infrastructure: %w", err)
	}
	ic.Print(logger)

	taskDef, err := ic.GetTaskDefinition()
	if err != nil {
		return fmt.Errorf("error retrieving current service information: %w", err)
	}

	logger.Printf("Deploying app %q\n", app.Name)
	taskDefArn := *taskDef.TaskDefinitionArn
	if imageTag := userConfig["imageTag"]; imageTag != "" {
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
	logger.Printf("Deployed app %q\n", app.Name)
	return nil
}
