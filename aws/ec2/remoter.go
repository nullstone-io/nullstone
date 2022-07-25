package ec2

import (
	"context"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	"gopkg.in/nullstone-io/nullstone.v0/aws/ssm"
	"gopkg.in/nullstone-io/nullstone.v0/config"
)

func NewRemoter(osWriters logging.OsWriters, nsConfig api.Config, appDetails app.Details) (admin.Remoter, error) {
	outs, err := outputs.Retrieve[Outputs](nsConfig, appDetails.Workspace)
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

func (r Remoter) Exec(ctx context.Context, task string, cmd string) error {
	return ExecCommand(ctx, r.Infra, cmd, nil)
}

func (r Remoter) Ssh(ctx context.Context, task string, forwards []config.PortForward) error {
	parameters, err := ssm.SessionParametersFromPortForwards(forwards)
	if err != nil {
		return err
	}

	return ExecCommand(ctx, r.Infra, "/bin/sh", parameters)
}
