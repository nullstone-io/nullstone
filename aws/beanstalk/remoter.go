package beanstalk

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	"gopkg.in/nullstone-io/nullstone.v0/aws/ssm"
)

func NewRemoter(ctx context.Context, osWriters logging.OsWriters, source outputs.RetrieverSource, appDetails app.Details) (admin.Remoter, error) {
	outs, err := outputs.Retrieve[Outputs](ctx, source, appDetails.Workspace, appDetails.WorkspaceConfig)
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
	instanceId, err := r.getInstanceId(ctx, options)
	if err != nil {
		return err
	}
	return ssm.StartEc2Session(ctx, r.Infra.AdminerConfig(), r.Infra.Region, instanceId, nil)
}

func (r Remoter) Ssh(ctx context.Context, options admin.RemoteOptions) error {
	parameters, err := ssm.SessionParametersFromPortForwards(options.PortForwards)
	if err != nil {
		return err
	}

	instanceId, err := r.getInstanceId(ctx, options)
	if err != nil {
		return err
	}
	return ssm.StartEc2Session(ctx, r.Infra.AdminerConfig(), r.Infra.Region, instanceId, parameters)
}

func (r Remoter) getInstanceId(ctx context.Context, options admin.RemoteOptions) (string, error) {
	if options.Instance == "" {
		if instanceId, err := GetRandomInstance(ctx, r.Infra); err != nil {
			return "", err
		} else if instanceId == "" {
			return "", fmt.Errorf("cannot exec command with no running instances")
		} else {
			return instanceId, nil
		}
	}
	return options.Instance, nil
}
