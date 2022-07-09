package ecr

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	dockertypes "github.com/docker/docker/api/types"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/aws"
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws-ecr"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
	"gopkg.in/nullstone-io/nullstone.v0/outputs"
	"log"
	"strings"
)

func NewPusher(logger *log.Logger, nsConfig api.Config, appDetails app.Details) (*Pusher, error) {
	outs, err := outputs.Retrieve[aws_ecr.Outputs](nsConfig, appDetails.Workspace)
	if err != nil {
		return nil, err
	}
	return &Pusher{
		Logger:   logger,
		NsConfig: nsConfig,
		Infra:    outs,
	}, nil
}

type Pusher struct {
	Logger   *log.Logger
	NsConfig api.Config
	Infra    aws_ecr.Outputs
}

func (p Pusher) Push(ctx context.Context, source, version string) error {
	// TODO: Log information to logger

	sourceUrl := docker.ParseImageUrl(source)
	targetUrl := p.Infra.ImageRepoUrl
	targetUrl.Tag = version

	if err := p.validate(targetUrl); err != nil {
		return err
	}

	targetAuth, err := p.getEcrLoginAuth(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving image registry credentials: %w", err)
	}

	p.Logger.Printf("Retagging %s => %s\n", sourceUrl.String(), targetUrl.String())
	if err := p.retagImage(ctx, sourceUrl, targetUrl); err != nil {
		return fmt.Errorf("error retagging image: %w", err)
	}

	p.Logger.Printf("Pushing %s\n", targetUrl.String())
	if err := docker.PushImage(ctx, targetUrl, targetAuth); err != nil {
		return fmt.Errorf("error pushing image: %w", err)
	}

	return nil
}

func (p Pusher) validate(targetUrl docker.ImageUrl) error {
	if targetUrl.String() == "" {
		return fmt.Errorf("cannot push if 'image_repo_url' module output is missing")
	}
	if targetUrl.Tag == "" {
		return fmt.Errorf("no version was specified, version is required to push image")
	}
	if !strings.Contains(targetUrl.Registry, "ecr") &&
		!strings.Contains(targetUrl.Registry, "amazonaws.com") {
		return fmt.Errorf("this app only supports push to AWS ECR (image=%s)", targetUrl)
	}

	// NOTE: For now, we are assuming that the production docker image is hosted in ECR
	// This will likely need to be refactored to support pushing to other image registries
	if p.Infra.ImagePusher.AccessKeyId == "" {
		return fmt.Errorf("cannot push without an authorized user, make sure 'image_pusher' output is not empty")
	}

	return nil
}

func (p Pusher) getEcrLoginAuth(ctx context.Context) (dockertypes.AuthConfig, error) {
	ecrClient := ecr.NewFromConfig(nsaws.NewConfig(p.Infra.ImagePusher, p.Infra.Region))
	out, err := ecrClient.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return dockertypes.AuthConfig{}, err
	}
	if len(out.AuthorizationData) > 0 {
		authData := out.AuthorizationData[0]
		token, err := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
		if err != nil {
			return dockertypes.AuthConfig{}, fmt.Errorf("invalid authorization token: %w", err)
		}
		tokens := strings.SplitN(string(token), ":", 2)
		return dockertypes.AuthConfig{
			Username:      tokens[0],
			Password:      tokens[1],
			ServerAddress: *authData.ProxyEndpoint,
		}, nil
	}
	return dockertypes.AuthConfig{}, nil
}

func (p Pusher) retagImage(ctx context.Context, sourceUrl, targetUrl docker.ImageUrl) error {
	dockerClient, err := docker.DiscoverDockerClient()
	if err != nil {
		return fmt.Errorf("error creating docker client: %w", err)
	}
	return dockerClient.ImageTag(ctx, sourceUrl.String(), targetUrl.String())
}
