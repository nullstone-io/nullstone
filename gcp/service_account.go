package gcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type ServiceAccount struct {
	Email      string `json:"email"`
	PrivateKey string `json:"private_key"`
}

func (a ServiceAccount) TokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	decoded, err := base64.StdEncoding.DecodeString(a.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("service account private key is not base64-encoded: %w", err)
	}
	cfg, err := google.JWTConfigFromJSON(decoded)
	if err != nil {
		return nil, fmt.Errorf("unable to read service account credentials json file: %w", err)
	}
	return cfg.TokenSource(ctx), nil
}
