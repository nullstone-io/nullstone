package ecs

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/go-multierror"
	nsaws "github.com/nullstone-io/deployment-sdk/aws"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
)

func RunTask(ctx context.Context, infra Outputs, options admin.RemoteOptions, cmd []string) error {
	ecsClient := ecs.NewFromConfig(nsaws.NewConfig(infra.Deployer, infra.Region))
	input := &ecs.RunTaskInput{
		TaskDefinition:       aws.String(infra.TaskArn),
		Cluster:              aws.String(infra.ClusterArn()),
		Count:                aws.Int32(1),
		EnableECSManagedTags: true,
		EnableExecuteCommand: true,
		LaunchType:           "FARGATE",
		Overrides:            &types.TaskOverride{},
		PropagateTags:        types.PropagateTagsService,
		StartedBy:            aws.String("nullstone-cli"), // TODO: Set this to nullstone-cli/<cur-user>
	}

	if infra.AppSecurityGroupId != "" && len(infra.TaskSubnetIds) > 0 {
		input.NetworkConfiguration = &types.NetworkConfiguration{
			AwsvpcConfiguration: &types.AwsVpcConfiguration{
				AssignPublicIp: types.AssignPublicIpDisabled,
				Subnets:        infra.TaskSubnetIds,
				SecurityGroups: []string{infra.AppSecurityGroupId},
			},
		}
	}

	if len(cmd) > 0 {
		input.Overrides.ContainerOverrides = []types.ContainerOverride{{Command: cmd}}
	}

	out, err := ecsClient.RunTask(ctx, input)
	if err != nil {
		return err
	}

	var errs error
	for _, failure := range out.Failures {
		err := fmt.Errorf("task (%s) failure: %s (%s)", aws.ToString(failure.Arn), aws.ToString(failure.Reason), aws.ToString(failure.Detail))
		errs = multierror.Append(errs, err)
	}
	if errs != nil {
		return errs
	}

	// TODO: Should we wait for the task to finish or rely on the user to track it?

	return nil
}
