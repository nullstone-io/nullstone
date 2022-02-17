package ssm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func StartEcsSession(session *ecstypes.Session, region, cluster, task, containerName string) error {
	targetRaw, _ := json.Marshal(ssm.StartSessionInput{
		Target: aws.String(fmt.Sprintf("ecs:%s_%s_%s", cluster, task, containerName)),
	})

	er := ecs.NewDefaultEndpointResolver()
	endpoint, _ := er.ResolveEndpoint(region, ecs.EndpointResolverOptions{})

	return StartSession(context.Background(), session, region, string(targetRaw), endpoint.URL)
}
