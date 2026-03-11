package vm

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
	return RunCommand(ctx, r.Infra, cmd)
}

func (r Remoter) Ssh(ctx context.Context, options admin.RemoteOptions) error {
	if r.Infra.BastionHostName == "" {
		return fmt.Errorf("interactive SSH requires an Azure Bastion host; ensure `bastion_host_name` is configured in module outputs")
	}
	return StartBastionSsh(ctx, r.Infra, options.Username)
}

func (r Remoter) Run(ctx context.Context, options admin.RunOptions, cmd []string, envVars map[string]string) error {
	return fmt.Errorf("`run` is not supported for Azure Virtual Machines")
}
