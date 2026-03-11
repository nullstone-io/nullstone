package aks

import (
	"context"

	"github.com/nullstone-io/deployment-sdk/k8s"
	"gopkg.in/nullstone-io/nullstone.v0/azure"
	"k8s.io/client-go/rest"
)

func CreateKubeConfig(ctx context.Context, cluster k8s.ClusterInfoer, principal azure.Principal) (*rest.Config, error) {
	configCreator := &k8s.ConfigCreator{
		ClusterInfoer: cluster,
		AuthInfoer:    PrincipalAuth{Principal: principal},
	}
	return configCreator.Create(ctx)
}
