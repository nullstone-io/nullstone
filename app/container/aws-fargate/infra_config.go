package aws_fargate

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	nsaws "gopkg.in/nullstone-io/nullstone.v0/aws"
	aws_fargate_service "gopkg.in/nullstone-io/nullstone.v0/contracts/aws-fargate-service"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
	"log"
)

// InfraConfig provides a minimal understanding of the infrastructure provisioned for a module type=aws-fargate
type InfraConfig struct {
	Outputs aws_fargate_service.Outputs
}

func (c InfraConfig) Print(logger *log.Logger) {
	logger.Printf("Using fargate cluster %q\n", c.Outputs.Cluster.ClusterArn)
	logger.Printf("Using fargate service %q\n", c.Outputs.ServiceName)
}

func (c InfraConfig) GetTaskDefinition() (*ecstypes.TaskDefinition, error) {
	client := ecs.NewFromConfig(nsaws.NewConfig(c.Outputs.Cluster.Deployer))

	out1, err := client.DescribeServices(context.Background(), &ecs.DescribeServicesInput{
		Services: []string{c.Outputs.ServiceName},
		Cluster:  aws.String(c.Outputs.Cluster.ClusterArn),
	})
	if err != nil {
		return nil, err
	}
	if len(out1.Services) < 1 {
		return nil, fmt.Errorf("could not find service %q in cluster %q", c.Outputs.ServiceName, c.Outputs.Cluster.ClusterArn)
	}

	out2, err := client.DescribeTaskDefinition(context.Background(), &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: out1.Services[0].TaskDefinition,
	})
	if err != nil {
		return nil, err
	}
	return out2.TaskDefinition, nil
}

func (c InfraConfig) UpdateTaskImageTag(taskDefinition *ecstypes.TaskDefinition, imageTag string) (*ecstypes.TaskDefinition, error) {
	client := ecs.NewFromConfig(nsaws.NewConfig(c.Outputs.Cluster.Deployer))

	defIndex, err := findMainContainerDefinitionIndex(taskDefinition.ContainerDefinitions)
	if err != nil {
		return nil, err
	}

	existingImageUrl := docker.ParseImageUrl(*taskDefinition.ContainerDefinitions[defIndex].Image)
	existingImageUrl.Digest = ""
	existingImageUrl.Tag = imageTag
	taskDefinition.ContainerDefinitions[defIndex].Image = aws.String(existingImageUrl.String())

	out, err := client.RegisterTaskDefinition(context.Background(), &ecs.RegisterTaskDefinitionInput{
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

	_, err = client.DeregisterTaskDefinition(context.Background(), &ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: taskDefinition.TaskDefinitionArn,
	})
	if err != nil {
		return nil, err
	}

	return out.TaskDefinition, nil
}

func findMainContainerDefinitionIndex(containerDefs []ecstypes.ContainerDefinition) (int, error) {
	mainIndex := -1
	for i, cd := range containerDefs {
		if cd.Essential != nil && *cd.Essential {
			if mainIndex > -1 {
				return 0, fmt.Errorf("cannot deploy a service with multiple containers marked as essential")
			}
			mainIndex = i
		}
	}
	if mainIndex > -1 {
		return mainIndex, nil
	}

	if len(containerDefs) == 0 {
		return 0, fmt.Errorf("cannot deploy service with no container definitions")
	}
	if len(containerDefs) > 1 {
		return 0, fmt.Errorf("cannot deploy service with multiple container definitions unless a single is marked essential")
	}
	return 0, nil
}

func (c InfraConfig) UpdateServiceTask(taskDefinitionArn string) error {
	client := ecs.NewFromConfig(nsaws.NewConfig(c.Outputs.Cluster.Deployer))

	_, err := client.UpdateService(context.Background(), &ecs.UpdateServiceInput{
		Service:            aws.String(c.Outputs.ServiceName),
		Cluster:            aws.String(c.Outputs.Cluster.ClusterArn),
		ForceNewDeployment: true,
		TaskDefinition:     aws.String(taskDefinitionArn),
	})
	return err
}
