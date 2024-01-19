package gke

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	"gopkg.in/nullstone-io/nullstone.v0/k8s"
	"os"
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
	opts := &k8s.ExecOptions{
		In:     os.Stdin,
		Out:    r.OsWriters.Stdout(),
		ErrOut: r.OsWriters.Stderr(),
		TTY:    false,
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
	if len(options.PortForwards) > 0 {
		return fmt.Errorf("gke provider does not support port forwarding yet")
	}

	return ExecCommand(ctx, r.Infra, options.Pod, options.Container, []string{"/bin/sh"}, opts)
}
