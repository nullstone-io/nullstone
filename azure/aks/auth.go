package aks

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/nullstone-io/deployment-sdk/azure"
	"github.com/nullstone-io/deployment-sdk/k8s"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var _ k8s.AuthInfoer = PrincipalAuth{}

// PrincipalAuth implements k8s.AuthInfoer using an Azure Principal's access token
type PrincipalAuth struct {
	azure.Principal
}

func (a PrincipalAuth) AuthInfo(ctx context.Context) (clientcmdapi.AuthInfo, error) {
	token, err := a.GetToken(ctx, policy.TokenRequestOptions{})
	if err != nil {
		return clientcmdapi.AuthInfo{}, fmt.Errorf("error retrieving kubernetes access token from Azure: %w", err)
	}
	return clientcmdapi.AuthInfo{
		Token: token.Token,
	}, nil
}
