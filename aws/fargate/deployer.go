package fargate

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/aws/ecs"
	"gopkg.in/nullstone-io/nullstone.v0/generic"
	"io"
	"os"
)

// Deployer will deploy an app/container using the following workflow to a fargate cluster:
//   Get task definition
//   Change image tag in task definition
//   Register new task definition
//   Deregister old task definition
//   Update ECS Service
type Deployer struct{}

func (d Deployer) Deploy(workspace *types.Workspace, imageTag string) error {
	deployContext, err := d.collectContext(workspace)
	if err != nil {
		return err
	}
	deployContext.Print(os.Stderr)

	taskDef, err := ecs.GetTaskDefinitionByServiceInCluster(deployContext.AwsConfig, deployContext.ClusterArn, deployContext.ServiceName)
	if err != nil {
		return fmt.Errorf("error retrieving current service information: %w", err)
	}

	taskDefArn := *taskDef.TaskDefinitionArn
	if imageTag != "" {
		fmt.Fprintf(os.Stderr, "Updating image tag to %q\n", imageTag)
		newTaskDef, err := ecs.UpdateTaskImageTag(deployContext.AwsConfig, taskDef, imageTag)
		if err != nil {
			return fmt.Errorf("error updating task with new image tag: %w", err)
		}
		taskDefArn = *newTaskDef.TaskDefinitionArn
	}

	if err := ecs.UpdateServiceTask(deployContext.AwsConfig, deployContext.ClusterArn, deployContext.ServiceName, taskDefArn); err != nil {
		return fmt.Errorf("error deploying service: %w", err)
	}
	return nil
}

type deployContext struct {
	ClusterArn  string
	ServiceName string
	AwsConfig   aws.Config
}

func (c deployContext) Print(w io.Writer) {
	fmt.Fprintf(w, "Using fargate cluster %q\n", c.ClusterArn)
	fmt.Fprintf(w, "Using fargate service %q\n", c.ServiceName)
}

func (d Deployer) collectContext(workspace *types.Workspace) (*deployContext, error) {
	if workspace.LastSuccessfulRun == nil || workspace.LastSuccessfulRun.Apply == nil {
		return nil, fmt.Errorf("cannot find outputs for application")
	}

	apply := workspace.LastSuccessfulRun.Apply
	dc := &deployContext{
		ClusterArn:  "",
		ServiceName: "",
	}
	missing := generic.ErrMissingOutputs{OutputNames: []string{}}

	clusterArnItem, ok := apply.Outputs["cluster_arn"]
	if !ok {
		missing.OutputNames = append(missing.OutputNames, "cluster_arn")
	} else {
		dc.ClusterArn, _ = clusterArnItem.Value.(string)
	}

	serviceNameItem, ok := apply.Outputs["service_name"]
	if !ok {
		missing.OutputNames = append(missing.OutputNames, "service_name")
	} else {
		dc.ServiceName, _ = serviceNameItem.Value.(string)
	}

	if len(missing.OutputNames) > 0 {
		return nil, missing
	}

	// TODO: Load deployer user from outputs
	var err error
	dc.AwsConfig, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		return nil, err
	}

	return dc, nil
}
