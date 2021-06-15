package aws_ecr

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/docker/docker/api/types"
	nsaws "gopkg.in/nullstone-io/nullstone.v0/aws"
	aws_ecr "gopkg.in/nullstone-io/nullstone.v0/contracts/aws-ecr"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
	"log"
	"strings"
)

// InfraConfig provides a minimal understanding of the infrastructure provisioned for a module type=aws-fargate
type InfraConfig struct {
	Outputs aws_ecr.Outputs
}

func (c InfraConfig) Print(logger *log.Logger) {
	logger = log.New(logger.Writer(), "    ", 0)
	logger.Printf("repository image url: %q\n", c.Outputs.ImageRepoUrl)
}

func (c InfraConfig) GetEcrLoginAuth() (types.AuthConfig, error) {
	ecrClient := ecr.NewFromConfig(nsaws.NewConfig(c.Outputs.ImagePusher, c.Outputs.Region))
	out, err := ecrClient.GetAuthorizationToken(context.TODO(), &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return types.AuthConfig{}, err
	}
	if len(out.AuthorizationData) > 0 {
		authData := out.AuthorizationData[0]
		token, err := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
		if err != nil {
			return types.AuthConfig{}, fmt.Errorf("invalid authorization token: %w", err)
		}
		tokens := strings.SplitN(string(token), ":", 2)
		return types.AuthConfig{
			Username:      tokens[0],
			Password:      tokens[1],
			ServerAddress: *authData.ProxyEndpoint,
		}, nil
	}
	return types.AuthConfig{}, nil
}

func (c InfraConfig) RetagImage(ctx context.Context, sourceUrl, targetUrl docker.ImageUrl) error {
	dockerClient, err := docker.DiscoverDockerClient()
	if err != nil {
		return fmt.Errorf("error creating docker client: %w", err)
	}
	return dockerClient.ImageTag(ctx, sourceUrl.String(), targetUrl.String())
}

func (c InfraConfig) PushImage(ctx context.Context, targetUrl docker.ImageUrl, targetAuth types.AuthConfig) error {
	return docker.PushImage(ctx, targetUrl, targetAuth)
}
