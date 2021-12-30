package nsaws

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/docker/docker/api/types"
	"gopkg.in/nullstone-io/nullstone.v0/contracts/aws"
	"strings"
)

func GetEcrLoginAuth(ctx context.Context, imagePusher aws.User, region string) (types.AuthConfig, error) {
	ecrClient := ecr.NewFromConfig(NewConfig(imagePusher, region))
	out, err := ecrClient.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
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
