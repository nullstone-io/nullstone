package cloudrun

import (
	"context"
	"fmt"

	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
)

var (
	_ admin.Remoter = Remoter{}
)

func NewRemoter(ctx context.Context, osWriters logging.OsWriters, source outputs.RetrieverSource, appDetails app.Details) (admin.Remoter, error) {
	outs, err := outputs.Retrieve[Outputs](ctx, source, appDetails.Workspace, appDetails.WorkspaceConfig)
	if err != nil {
		return nil, err
	}
	outs.InitializeCreds(source, appDetails.Workspace)

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
	if r.Infra.ServiceName == "" {
		return fmt.Errorf("cannot `exec` unless you have a long-running service, use `run` for a job/task")
	}

	// TODO: Implement
	return nil
}

func (r Remoter) Ssh(ctx context.Context, options admin.RemoteOptions) error {
	if r.Infra.ServiceName == "" {
		return fmt.Errorf("cannot `ssh` unless you have a long-running service, use `run` for a job/task")
	}

	// TODO: Implement
	return nil
}

func (r Remoter) Run(ctx context.Context, options admin.RunOptions, cmd []string, envVars map[string]string) error {
	if r.Infra.ServiceName != "" {
		return fmt.Errorf("cannot use `run` for a long-running service, use `exec` instead")
	}

	runner := JobRunner{
		JobName:           r.Infra.JobName,
		MainContainerName: r.Infra.MainContainerName,
		Adminer:           r.Infra.Deployer,
	}
	return runner.Run(ctx, options, cmd, envVars)
}
