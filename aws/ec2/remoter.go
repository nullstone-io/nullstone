package ec2

import (
	"context"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	"gopkg.in/nullstone-io/nullstone.v0/aws/ssm"
)

func NewRemoter(osWriters logging.OsWriters, source outputs.RetrieverSource, appDetails app.Details) (admin.Remoter, error) {
	outs, err := outputs.Retrieve[Outputs](source, appDetails.Workspace)
	if err != nil {
		return nil, err
	}

	return Remoter{
		OsWriters: osWriters,
		Details:   appDetails,
		Infra:     outs,
	}, nil
}

type Remoter struct {
	OsWriters logging.OsWriters
	Details   app.Details
	Infra     Outputs
}

func (r Remoter) Exec(ctx context.Context, options admin.RemoteOptions, cmd []string) error {
	// TODO: Add support for cmd
	return ssm.StartEc2Session(ctx, r.Infra.AdminerConfig(), r.Infra.Region, r.Infra.InstanceId, nil)
}

func (r Remoter) Ssh(ctx context.Context, options admin.RemoteOptions) error {
	parameters, err := ssm.SessionParametersFromPortForwards(options.PortForwards)
	if err != nil {
		return err
	}

	return ssm.StartEc2Session(ctx, r.Infra.AdminerConfig(), r.Infra.Region, r.Infra.InstanceId, parameters)
}
