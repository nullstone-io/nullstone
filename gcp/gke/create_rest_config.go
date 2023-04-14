package gke

import (
	"context"
	"github.com/nullstone-io/deployment-sdk/gcp"
	"github.com/nullstone-io/deployment-sdk/gcp/gke"
	"github.com/nullstone-io/deployment-sdk/k8s"
	"k8s.io/client-go/rest"
)

func CreateRESTConfig(ctx context.Context, serviceAccount gcp.ServiceAccount, cluster k8s.ClusterInfoer) (*rest.Config, error) {
	configCreator := &k8s.ConfigCreator{
		TokenSourcer:  serviceAccount,
		ClusterInfoer: cluster,
	}
	return configCreator.Create(ctx, gke.GcpScopes...)
}
