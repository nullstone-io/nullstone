package lambda

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	nslambda "github.com/nullstone-io/deployment-sdk/aws/lambda"
)

func Invoke(ctx context.Context, infra nslambda.Outputs, payload []byte, async bool) error {
	λClient := lambda.NewFromConfig(infra.DeployerAwsConfig())
	input := &lambda.InvokeInput{
		FunctionName:   aws.String(infra.FunctionName()),
		InvocationType: types.InvocationTypeRequestResponse,
		LogType:        types.LogTypeTail,
		Payload:        payload,
	}
	if async {
		input.InvocationType = types.InvocationTypeEvent
		input.LogType = types.LogTypeNone
	}
	out, err := λClient.Invoke(ctx, input)
	if err != nil {
		return err
	}

	out.LogResult
}
