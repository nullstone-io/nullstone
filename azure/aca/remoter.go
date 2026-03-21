package aca

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
	if r.Infra.ContainerAppName == "" {
		return fmt.Errorf("cannot `exec` unless you have a long-running container app, use `run` for a job")
	}
	if len(options.PortForwards) > 0 {
		return fmt.Errorf("Azure Container Apps does not support port forwarding")
	}

	return ExecCommand(ctx, r.Infra, options.Container, cmd)
}

func (r Remoter) Ssh(ctx context.Context, options admin.RemoteOptions) error {
	if r.Infra.ContainerAppName == "" {
		return fmt.Errorf("cannot `ssh` unless you have a long-running container app")
	}
	if len(options.PortForwards) > 0 {
		return fmt.Errorf("Azure Container Apps does not support port forwarding")
	}

	return ExecCommand(ctx, r.Infra, options.Container, []string{"/bin/sh"})
}

func (r Remoter) Run(ctx context.Context, options admin.RunOptions, cmd []string, envVars map[string]string) error {
	if r.Infra.ContainerAppName != "" {
		return fmt.Errorf("cannot use `run` with a long-running container app, use `exec` instead")
	}
	if r.Infra.JobName == "" {
		return fmt.Errorf("no job name configured for this application")
	}

	return RunJob(ctx, r.OsWriters, r.Infra, cmd, envVars)
}
