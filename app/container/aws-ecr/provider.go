package aws_ecr

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
	// NOTE: We expect --version from the user which is used as the image tag for the pushed image
	if imageTag := userConfig["version"]; imageTag != "" {
		targetUrl.Tag = imageTag
	} else {
		targetUrl.Tag = sourceUrl.Tag
	}
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
	fmt.Printf("%+v\n", targetAuth)

	logger.Printf("Retagging %s => %s\n", sourceUrl.String(), targetUrl.String())
	if err := ic.RetagImage(ctx, sourceUrl, targetUrl); err != nil {
		return fmt.Errorf("error retagging image: %w", err)
	}

	logger.Printf("Pushing %s\n", targetUrl.String())
	if err := ic.PushImage(ctx, targetUrl, targetAuth); err != nil {
		return fmt.Errorf("error pushing image: %w", err)
	}

	return nil
}

// Deploy is not supported for this provider
func (p Provider) Deploy(nsConfig api.Config, application *types.Application, workspace *types.Workspace, userConfig map[string]string) error {
	return fmt.Errorf("cannot deploy app/container apps defined with type service/aws-ecr")
}
