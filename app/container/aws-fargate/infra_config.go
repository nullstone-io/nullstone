package aws_fargate

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/docker/docker/api/types"
	nsaws "gopkg.in/nullstone-io/nullstone.v0/aws"
	aws_fargate_service "gopkg.in/nullstone-io/nullstone.v0/contracts/aws-fargate-service"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
	"log"
	"strings"
)

// InfraConfig provides a minimal understanding of the infrastructure provisioned for a module type=aws-fargate
type InfraConfig struct {
	Outputs aws_fargate_service.Outputs
}

func (c InfraConfig) Print(logger *log.Logger) {
	logger = log.New(logger.Writer(), "    ", 0)
	logger.Printf("fargate cluster: %q\n", c.Outputs.Cluster.ClusterArn)
	logger.Printf("fargate service: %q\n", c.Outputs.ServiceName)
	logger.Printf("repository image url: %q\n", c.Outputs.ImageRepoUrl)
}

func (c InfraConfig) GetTaskDefinition() (*ecstypes.TaskDefinition, error) {
	ecsClient := ecs.NewFromConfig(nsaws.NewConfig(c.Outputs.Cluster.Deployer, c.Outputs.Region))

	out1, err := ecsClient.DescribeServices(context.Background(), &ecs.DescribeServicesInput{
		Services: []string{c.Outputs.ServiceName},
		Cluster:  aws.String(c.Outputs.Cluster.ClusterArn),
	})
	if err != nil {
		return nil, err
	}
	if len(out1.Services) < 1 {
		return nil, fmt.Errorf("could not find service %q in cluster %q", c.Outputs.ServiceName, c.Outputs.Cluster.ClusterArn)
	}

	out2, err := ecsClient.DescribeTaskDefinition(context.Background(), &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: out1.Services[0].TaskDefinition,
	})
	if err != nil {
		return nil, err
	}
	return out2.TaskDefinition, nil
}

func (c InfraConfig) UpdateTaskImageTag(taskDefinition *ecstypes.TaskDefinition, imageTag string) (*ecstypes.TaskDefinition, error) {
	ecsClient := ecs.NewFromConfig(nsaws.NewConfig(c.Outputs.Cluster.Deployer, c.Outputs.Region))

	defIndex, err := c.findMainContainerDefinitionIndex(taskDefinition.ContainerDefinitions)
	if err != nil {
		return nil, err
	}

	existingImageUrl := docker.ParseImageUrl(*taskDefinition.ContainerDefinitions[defIndex].Image)
	existingImageUrl.Digest = ""
	existingImageUrl.Tag = imageTag
	taskDefinition.ContainerDefinitions[defIndex].Image = aws.String(existingImageUrl.String())

	out, err := ecsClient.RegisterTaskDefinition(context.Background(), &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions:    taskDefinition.ContainerDefinitions,
		Family:                  taskDefinition.Family,
		Cpu:                     taskDefinition.Cpu,
		ExecutionRoleArn:        taskDefinition.ExecutionRoleArn,
		InferenceAccelerators:   taskDefinition.InferenceAccelerators,
		IpcMode:                 taskDefinition.IpcMode,
		Memory:                  taskDefinition.Memory,
		NetworkMode:             taskDefinition.NetworkMode,
		PidMode:                 taskDefinition.PidMode,
		PlacementConstraints:    taskDefinition.PlacementConstraints,
		ProxyConfiguration:      taskDefinition.ProxyConfiguration,
		RequiresCompatibilities: taskDefinition.RequiresCompatibilities,
		TaskRoleArn:             taskDefinition.TaskRoleArn,
		Volumes:                 taskDefinition.Volumes,
	})
	if err != nil {
		return nil, err
	}

	_, err = ecsClient.DeregisterTaskDefinition(context.Background(), &ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: taskDefinition.TaskDefinitionArn,
	})
	if err != nil {
		return nil, err
	}

	return out.TaskDefinition, nil
}

func (c InfraConfig) findMainContainerDefinitionIndex(containerDefs []ecstypes.ContainerDefinition) (int, error) {
	if len(containerDefs) == 0 {
		return -1, fmt.Errorf("cannot deploy service with no container definitions")
	}
	if len(containerDefs) == 1 {
		return 0, nil
	}

	if mainContainerName := c.Outputs.MainContainerName; mainContainerName != "" {
		// let's go find main_container_name
		for i, cd := range containerDefs {
			if cd.Name != nil && *cd.Name == mainContainerName {
				return i, nil
			}
		}
		return -1, fmt.Errorf("cannot deploy service; no container definition with main_container_name = %s", mainContainerName)
	}

	// main_container_name was not specified, we are going to attempt to find a single container definition
	// If more than one container definition exists, we will error
	if len(containerDefs) > 1 {
		return -1, fmt.Errorf("service contains multiple containers; cannot deploy unless service module exports 'main_container_name'")
	}
	return 0, nil
}

func (c InfraConfig) GetService() (*ecstypes.Service, error) {
	ecsClient := ecs.NewFromConfig(nsaws.NewConfig(c.Outputs.Cluster.Deployer, c.Outputs.Region))
	out, err := ecsClient.DescribeServices(context.Background(), &ecs.DescribeServicesInput{
		Services: []string{c.Outputs.ServiceName},
		Cluster:  aws.String(c.Outputs.Cluster.ClusterArn),
	})
	if err != nil {
		return nil, err
	}
	if len(out.Services) > 0 {
		return &out.Services[0], nil
	}
	return nil, nil
}

func (c InfraConfig) GetTargetGroupHealth(targetGroupArn string) ([]elbv2types.TargetHealthDescription, error) {
	elbClient := elasticloadbalancingv2.NewFromConfig(nsaws.NewConfig(c.Outputs.Cluster.Deployer, c.Outputs.Region))
	out, err := elbClient.DescribeTargetHealth(context.Background(), &elasticloadbalancingv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(targetGroupArn),
	})
	if err != nil {
		return nil, err
	}
	return out.TargetHealthDescriptions, nil
}

func (c InfraConfig) UpdateServiceTask(taskDefinitionArn string) error {
	ecsClient := ecs.NewFromConfig(nsaws.NewConfig(c.Outputs.Cluster.Deployer, c.Outputs.Region))

	_, err := ecsClient.UpdateService(context.Background(), &ecs.UpdateServiceInput{
		Service:            aws.String(c.Outputs.ServiceName),
		Cluster:            aws.String(c.Outputs.Cluster.ClusterArn),
		ForceNewDeployment: true,
		TaskDefinition:     aws.String(taskDefinitionArn),
	})
	return err
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
