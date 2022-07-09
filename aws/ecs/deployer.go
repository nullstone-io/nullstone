package ecs

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws-ecs-service"
	"gopkg.in/nullstone-io/nullstone.v0/outputs"
	"log"
)

func NewDeployer(logger *log.Logger, nsConfig api.Config, appDetails app.Details) (app.Deployer, error) {
	outs, err := outputs.Retrieve[aws_ecs_service.Outputs](nsConfig, appDetails.Workspace)
	if err != nil {
		return nil, err
	}

	return Deployer{
		Logger:   logger,
		NsConfig: nsConfig,
		Details:  appDetails,
		Infra:    outs,
	}, nil
}

type Deployer struct {
	Logger   *log.Logger
	NsConfig api.Config
	Details  app.Details
	Infra    aws_ecs_service.Outputs
}

func (d Deployer) Print() {
	d.Logger.Printf("ecs cluster: %q\n", d.Infra.Cluster.ClusterArn)
	d.Logger.Printf("ecs service: %q\n", d.Infra.ServiceName)
	d.Logger.Printf("repository image url: %q\n", d.Infra.ImageRepoUrl)
}

// Deploy takes the following steps to deploy an AWS ECS service
//   Get task definition
//   Change image tag in task definition
//   Register new task definition
//   Deregister old task definition
//   Update ECS Service (This always causes deployment)
func (d Deployer) Deploy(ctx context.Context, version string) (*string, error) {
	d.Print()

	if version == "" {
		return nil, fmt.Errorf("no version specified, version is required to deploy")
	}

	d.Logger.Printf("Deploying app %q\n", d.Details.App.Name)

	taskDef, err := aws_ecs_service.GetTaskDefinition(ctx, d.Infra)
	if err != nil {
		return nil, fmt.Errorf("error retrieving current service information: %w", err)
	} else if taskDef == nil {
		return nil, fmt.Errorf("could not find task definition")
	}

	d.Logger.Printf("Updating image tag to %q\n", version)
	newTaskDef, err := aws_ecs_service.UpdateTaskImageTag(ctx, d.Infra, taskDef, version)
	if err != nil {
		return nil, fmt.Errorf("error updating task with new image tag: %w", err)
	}
	newTaskDefArn := *newTaskDef.TaskDefinitionArn

	if d.Infra.ServiceName == "" {
		d.Logger.Printf("No service name in outputs. Skipping update service.")
		return nil, nil
	}

	deployment, err := aws_ecs_service.UpdateServiceTask(ctx, d.Infra, newTaskDefArn)
	if err != nil {
		return nil, fmt.Errorf("error deploying service: %w", err)
	}
	d.Logger.Printf("Deployed app %q\n", d.Details.App.Name)
	return deployment.Id, nil
}
