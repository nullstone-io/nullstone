package azure

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// Principal contains credentials for an Azure identity that has access to perform a particular action
// This structure must match the fields defined in outputs of the module
type Principal struct {
	TenantId string `json:"tenant_id"`
	ClientId string `json:"client_id"`

	RemoteTokenProvider TokenProvider `json:"-"`
}

// TokenProvider retrieves an Azure access token from a remote source (e.g., Nullstone API)
type TokenProvider interface {
	GetToken(ctx context.Context) (string, error)
}

// GetToken retrieves an Azure access token, falling back to the remote provider
func (p Principal) GetToken(ctx context.Context) (string, error) {
	if p.RemoteTokenProvider != nil {
		return p.RemoteTokenProvider.GetToken(ctx)
	}
	return "", fmt.Errorf("missing Azure credentials")
}

// NullstoneTokenProvider retrieves Azure access tokens from the Nullstone API
type NullstoneTokenProvider struct {
	RetrieverSource outputs.RetrieverSource
	StackId         int64
	WorkspaceUid    uuid.UUID
	OutputNames     []string
}

func (p NullstoneTokenProvider) GetToken(ctx context.Context) (string, error) {
	input := api.GenerateCredentialsInput{
		Provider:    types.ProviderAzure,
		OutputNames: p.OutputNames,
	}
	creds, err := p.RetrieverSource.GetTemporaryCredentials(ctx, p.StackId, p.WorkspaceUid, input)
	if err != nil {
		return "", fmt.Errorf("error retrieving temporary credentials from Nullstone: %w", err)
	}
	if creds == nil {
		return "", fmt.Errorf("no credentials returned from Nullstone")
	}
	token, ok := creds.Data["access_token"]
	if !ok || token == "" {
		return "", fmt.Errorf("Azure access token not found in Nullstone credentials response")
	}
	return token, nil
}

// NewTokenProviderFactory creates a factory function for generating NullstoneTokenProviders
func NewTokenProviderFactory(source outputs.RetrieverSource, stackId int64, workspaceUid uuid.UUID) func(outputNames ...string) TokenProvider {
	return func(outputNames ...string) TokenProvider {
		return NullstoneTokenProvider{
			RetrieverSource: source,
			StackId:         stackId,
			WorkspaceUid:    workspaceUid,
			OutputNames:     outputNames,
		}
	}
}
