package gcp

import (
	gcpcreds "cloud.google.com/go/iam/credentials/apiv1"
	"context"
	"fmt"
	pbtypes "github.com/golang/protobuf/ptypes"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/iam/credentials/v1"
	"time"
)

type ServiceAccount struct {
	Name    string `json:"name"`
	KeyFile string `json:"key_file"`
}

func (s *ServiceAccount) GenerateAccessToken(lifetime time.Duration) (string, error) {
	ctx := context.Background()
	credsClient, err := gcpcreds.NewIamCredentialsClient(ctx, option.WithCredentialsJSON([]byte(s.KeyFile)))
	if err != nil {
		return "", fmt.Errorf("error creating GCP client: %w", err)
	}

	res, err := credsClient.GenerateAccessToken(ctx, &credentials.GenerateAccessTokenRequest{
		Name: s.Name,
		Scope: []string{
			"openid",
			"https://www.googleapis.com/auth/cloud-platform",
			"https://www.googleapis.com/auth/accounts.reauth",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Lifetime: pbtypes.DurationProto(lifetime),
	})
	if err != nil {
		return "", fmt.Errorf("error generating access token: %w", err)
	}
	return res.GetAccessToken(), nil
}
