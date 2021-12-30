package nsgcp

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"gopkg.in/nullstone-io/nullstone.v0/contracts/gcp"
)

var (
	GCRScopes = []string{"https://www.googleapis.com/auth/cloud-platform"}
)

func GetGcrLoginAuth(ctx context.Context, imagePusher gcp.ServiceAccount, registry string) (types.AuthConfig, error) {
	ts, err := imagePusher.TokenSource(ctx, GCRScopes...)
	if err != nil {
		return types.AuthConfig{}, fmt.Errorf("error creating access token source: %w", err)
	}
	token, err := ts.Token()
	if err != nil {
		return types.AuthConfig{}, fmt.Errorf("error retrieving access token: %w", err)
	}
	if token == nil || token.AccessToken == "" {
		return types.AuthConfig{}, nil
	}

	serverAddr := registry
	if serverAddr == "" {
		serverAddr = "gcr.io"
	}
	return types.AuthConfig{
		ServerAddress: serverAddr,
		Username:      "oauth2accesstoken",
		Password:      token.AccessToken,
	}, nil
}
