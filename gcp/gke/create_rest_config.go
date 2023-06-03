package gke

import (
	"context"
	"github.com/nullstone-io/deployment-sdk/gcp"
	"github.com/nullstone-io/deployment-sdk/gcp/gke"
	"github.com/nullstone-io/deployment-sdk/k8s"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

func CreateRESTConfig(ctx context.Context, serviceAccount gcp.ServiceAccount, cluster k8s.ClusterInfoer) (*rest.Config, error) {
	configCreator := &k8s.ConfigCreator{
		TokenSourcer:  serviceAccount,
		ClusterInfoer: cluster,
	}
	cfg, err := configCreator.Create(ctx, gke.GcpScopes...)
	if err != nil {
		return nil, err
	}
	cfg.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	cfg.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	return cfg, nil
}
