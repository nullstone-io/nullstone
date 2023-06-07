package gke

import (
	"context"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/gcp/gke"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	"gopkg.in/nullstone-io/nullstone.v0/k8s"
	"k8s.io/client-go/kubernetes"
)

func NewLogStreamer(osWriters logging.OsWriters, nsConfig api.Config, appDetails app.Details) (admin.LogStreamer, error) {
	outs, err := outputs.Retrieve[Outputs](nsConfig, appDetails.Workspace)
	if err != nil {
		return nil, err
	}

	return k8s.LogStreamer{
		OsWriters:    osWriters,
		Details:      appDetails,
		AppNamespace: outs.ServiceNamespace,
		AppName:      outs.ServiceName,
		NewClientFn: func(ctx context.Context) (*kubernetes.Clientset, error) {
			return gke.CreateKubeClient(ctx, outs.ClusterNamespace, outs.Deployer)
		},
	}, nil
}
