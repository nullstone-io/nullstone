package eks

import (
	"context"
	"fmt"
	"os"

	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/aws/eks"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	"gopkg.in/nullstone-io/nullstone.v0/k8s"
	"k8s.io/client-go/rest"
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

	opts := &k8s.ExecOptions{
		In:     os.Stdin,
		Out:    r.OsWriters.Stdout(),
		ErrOut: r.OsWriters.Stderr(),
		TTY:    false,
	}
	for _, pf := range options.PortForwards {
		opts.PortMappings = append(opts.PortMappings, fmt.Sprintf("%s:%s", pf.LocalPort, pf.RemotePort))
	}

	return ExecCommand(ctx, r.Infra, options.Pod, options.Container, cmd, opts)
}

func (r Remoter) Ssh(ctx context.Context, options admin.RemoteOptions) error {
	opts := &k8s.ExecOptions{
		In:     os.Stdin,
		Out:    r.OsWriters.Stdout(),
		ErrOut: r.OsWriters.Stderr(),
		TTY:    true,
	}
	for _, pf := range options.PortForwards {
		opts.PortMappings = append(opts.PortMappings, fmt.Sprintf("%s:%s", pf.LocalPort, pf.RemotePort))
	}

	return ExecCommand(ctx, r.Infra, options.Pod, options.Container, []string{"/bin/sh"}, opts)
}

func (r Remoter) Run(ctx context.Context, options admin.RunOptions, cmd []string, envVars map[string]string) error {
	if r.Infra.ServiceName != "" {
		return fmt.Errorf("cannot use `run` for a long-running service, use `exec` instead")
	}

	runner := k8s.JobRunner{
		Namespace:         r.Infra.ServiceNamespace,
		AppName:           r.Details.App.Name,
		MainContainerName: r.Infra.MainContainerName,
		JobDefinitionName: r.Infra.JobDefinitionName,
		NewConfigFn: func(ctx context.Context) (*rest.Config, error) {
			return eks.CreateKubeConfig(ctx, r.Infra.ClusterNamespace, r.Infra.Deployer)
		},
		Out:    r.OsWriters.Stdout(),
		ErrOut: r.OsWriters.Stderr(),
	}
	return runner.Run(ctx, options, cmd, envVars)
}
