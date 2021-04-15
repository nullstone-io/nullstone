package fargate

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	nsaws "gopkg.in/nullstone-io/nullstone.v0/aws"
	"gopkg.in/nullstone-io/nullstone.v0/deploy"
	"gopkg.in/nullstone-io/nullstone.v0/docker"
	"gopkg.in/nullstone-io/nullstone.v0/generic"
	"log"
)

const (
	ClusterModuleType = "cluster/aws-fargate"
)

var _ deploy.InfraConfig = InfraConfig{}

// InfraConfig provides the mechanism through which AWS actions are performed
type InfraConfig struct {
	ClusterArn  string
	ServiceName string
	AwsConfig   aws.Config
}

func newInfraConfig(nsConfig api.Config, workspace *types.Workspace) (*InfraConfig, error) {
	dc := &InfraConfig{}
	missingErr := generic.ErrMissingOutputs{OutputNames: []string{}}

	// We need to retrieve the cluster workspace to extract the cluster arn and deployer user
	clusterWorkspace, err := generic.GetConnectionWorkspace(nsConfig, workspace, "", ClusterModuleType)
	if err != nil {
		return nil, fmt.Errorf("error finding cluster for application: %w", err)
	}
	if clusterWorkspace == nil {
		return nil, fmt.Errorf("cannot find cluster for application")
	}
	if clusterWorkspace.LastSuccessfulRun == nil || clusterWorkspace.LastSuccessfulRun.Apply == nil {
		return nil, fmt.Errorf("outputs missing from cluster")
	}
	clusterOutputs := clusterWorkspace.LastSuccessfulRun.Apply.Outputs

	deployerUser := nsaws.DeployerUser{}
	if !generic.ExtractStructFromOutputs(clusterOutputs, "deployer", &deployerUser) {
		missingErr.OutputNames = append(missingErr.OutputNames, "deployer")
	}
	dc.AwsConfig = deployerUser.CreateConfig()
	if dc.ClusterArn = generic.ExtractStringFromOutputs(clusterOutputs, "cluster_arn"); dc.ClusterArn == "" {
		missingErr.OutputNames = append(missingErr.OutputNames, "cluster_arn")
	}

	if workspace.LastSuccessfulRun == nil || workspace.LastSuccessfulRun.Apply == nil {
		return nil, fmt.Errorf("cannot find outputs for application")
	}
	workspaceOutputs := workspace.LastSuccessfulRun.Apply.Outputs
	if dc.ServiceName = generic.ExtractStringFromOutputs(workspaceOutputs, "service_name"); dc.ServiceName == "" {
		missingErr.OutputNames = append(missingErr.OutputNames, "service_name")
	}

	if len(missingErr.OutputNames) > 0 {
		return nil, missingErr
	}
	return dc, nil
}

func (c InfraConfig) Print(logger *log.Logger) {
	logger.Printf("Using fargate cluster %q\n", c.ClusterArn)
	logger.Printf("Using fargate service %q\n", c.ServiceName)
}

func (c InfraConfig) GetTaskDefinition() (*ecstypes.TaskDefinition, error) {
	client := ecs.NewFromConfig(c.AwsConfig)

	out1, err := client.DescribeServices(context.Background(), &ecs.DescribeServicesInput{
		Services: []string{c.ServiceName},
		Cluster:  aws.String(c.ClusterArn),
	})
	if err != nil {
		return nil, err
	}
	if len(out1.Services) < 1 {
		return nil, fmt.Errorf("could not find service %q in cluster %q", c.ServiceName, c.ClusterArn)
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
	client := ecs.NewFromConfig(c.AwsConfig)

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
	client := ecs.NewFromConfig(c.AwsConfig)

	_, err := client.UpdateService(context.Background(), &ecs.UpdateServiceInput{
		Service:            aws.String(c.ServiceName),
		Cluster:            aws.String(c.ClusterArn),
		ForceNewDeployment: true,
		TaskDefinition:     aws.String(taskDefinitionArn),
	})
	return err
}
