package ssm

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func StartEcsSession(ctx context.Context, config aws.Config, region, cluster, taskId, containerName, cmd string, parameters map[string][]string) error {
	docName := GetDocumentName(parameters)

	ecsClient := ecs.NewFromConfig(config)
	input := &ecs.ExecuteCommandInput{
		Cluster:     aws.String(cluster),
		Task:        aws.String(taskId),
		Container:   aws.String(containerName), // TODO: Allow user to select which container
		Command:     aws.String(cmd),
		Interactive: true,
	}
	out, err := ecsClient.ExecuteCommand(context.Background(), input)
	if err != nil {
		return fmt.Errorf("error establishing ecs execute command: %w", err)
	}

	target := ssm.StartSessionInput{
		DocumentName: docName,
		Target:       aws.String(fmt.Sprintf("ecs:%s_%s_%s", cluster, taskId, containerName)),
		Parameters:   parameters,
	}

	er := ecs.NewDefaultEndpointResolver()
	endpoint, _ := er.ResolveEndpoint(region, ecs.EndpointResolverOptions{})

	return StartSession(ctx, out.Session, target, region, endpoint.URL)
}
