package aws_fargate

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
	"gopkg.in/nullstone-io/nullstone.v0/outputs"
	"log"
	"os"
	"strings"
)

var (
	logger = log.New(os.Stderr, "", 0)
)

var _ app.Provider = Provider{}

type Provider struct {
}

func (p Provider) identify(nsConfig api.Config, app *types.Application, workspace *types.Workspace) (*InfraConfig, error) {
	logger.Printf("Identifying infrastructure for app %q\n", app.Name)
	ic := &InfraConfig{}
	retriever := outputs.Retriever{NsConfig: nsConfig}
	if err := retriever.Retrieve(workspace, &ic.Outputs); err != nil {
		return nil, fmt.Errorf("Unable to identify app infrastructure: %w", err)
	}
	ic.Print(logger)
	return ic, nil
}

func (p Provider) Push(nsConfig api.Config, app *types.Application, workspace *types.Workspace, userConfig map[string]string) error {
	ic, err := p.identify(nsConfig, app, workspace)
	if err != nil {
		return err
	}

	sourceUrl := docker.ParseImageUrl(userConfig["source"])

	targetUrl := ic.Outputs.ImageRepoUrl
	targetUrl.Tag = userConfig["imageTag"]
	if targetUrl.String() == "" {
		return fmt.Errorf("cannot push if 'image_repo_url' module output is missing")
	}
	if !strings.Contains(targetUrl.Registry, "ecr") &&
		!strings.Contains(targetUrl.Registry, "amazonaws.com") {
		return fmt.Errorf("this app only supports push to AWS ECR (image=%s)", targetUrl)
	}
	// NOTE: For now, we are assuming that the production docker image is hosted in ECR
	// This will likely need to be refactored to support pushing to other image registries
	if ic.Outputs.ImagePusher.AccessKeyId == "" {
		return fmt.Errorf("cannot push without an authorized user, make sure 'image_pusher' output is not empty")
	}

	// TODO: Add cancellation support so users can press Control+C to kill push
	ctx := context.TODO()

	targetAuth, err := ic.GetEcrLoginAuth()
	if err != nil {
		return fmt.Errorf("error retrieving image registry credentials: %w", err)
	}

	if err := ic.RetagImage(ctx, sourceUrl, targetUrl); err != nil {
		return fmt.Errorf("error retagging image: %w", err)
	}

	if err := ic.PushImage(ctx, targetUrl, targetAuth); err != nil {
		return fmt.Errorf("error pushing image: %w", err)
	}

	return nil
}

// Deploy takes the following steps to deploy an AWS Fargate service
//   Get task definition
//   Change image tag in task definition
//   Register new task definition
//   Deregister old task definition
//   Update ECS Service (This always causes deployment)
func (p Provider) Deploy(nsConfig api.Config, application *types.Application, workspace *types.Workspace, userConfig map[string]string) error {
	ic, err := p.identify(nsConfig, application, workspace)
	if err != nil {
		return err
	}

	taskDef, err := ic.GetTaskDefinition()
	if err != nil {
		return fmt.Errorf("error retrieving current service information: %w", err)
	}

	logger.Printf("Deploying app %q\n", application.Name)
	version := userConfig["version"]
	taskDefArn := *taskDef.TaskDefinitionArn
	if version != "" {
		fmt.Fprintf(os.Stderr, "Updating app version to %q\n", version)
		if err := app.UpdateVersion(nsConfig, application.Name, workspace.EnvName, version); err != nil {
			return fmt.Errorf("error updating app version in nullstone: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Updating image tag to %q\n", version)
		newTaskDef, err := ic.UpdateTaskImageTag(taskDef, version)
		if err != nil {
			return fmt.Errorf("error updating task with new image tag: %w", err)
		}
		taskDefArn = *newTaskDef.TaskDefinitionArn
	}

	if err := ic.UpdateServiceTask(taskDefArn); err != nil {
		return fmt.Errorf("error deploying service: %w", err)
	}

	logger.Printf("Deployed app %q\n", application.Name)
	return nil
}
