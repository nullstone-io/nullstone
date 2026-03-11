package aks

import (
	"context"
	"fmt"

	"github.com/nullstone-io/deployment-sdk/k8s"
	"gopkg.in/nullstone-io/nullstone.v0/azure"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var _ k8s.AuthInfoer = PrincipalAuth{}

// PrincipalAuth implements k8s.AuthInfoer using an Azure Principal's access token
type PrincipalAuth struct {
	azure.Principal
}

func (a PrincipalAuth) AuthInfo(ctx context.Context) (clientcmdapi.AuthInfo, error) {
	token, err := a.GetToken(ctx)
	if err != nil {
		return clientcmdapi.AuthInfo{}, fmt.Errorf("error retrieving kubernetes access token from Azure: %w", err)
	}
	return clientcmdapi.AuthInfo{
		Token: token,
	}, nil
}
