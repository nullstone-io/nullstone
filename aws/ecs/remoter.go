package ecs

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
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

func (r Remoter) Exec(ctx context.Context, options admin.RemoteOptions, cmd []string) error {
	if r.Infra.ServiceName == "" {
		// If there is no ServiceName, this is a task that is invokable
		return RunTask(ctx, r.Infra, options, cmd)
	}
	taskId, err := r.getTaskId(ctx, options)
	if err != nil {
		return err
	}
	return ExecCommand(ctx, r.Infra, taskId, options.Container, cmd, nil)
}

func (r Remoter) Ssh(ctx context.Context, options admin.RemoteOptions) error {
	taskId, err := r.getTaskId(ctx, options)
	if err != nil {
		return err
	}
	if len(options.PortForwards) > 0 {
		return fmt.Errorf("ecs provider does not support port forwarding")
	}
	return ExecCommand(ctx, r.Infra, taskId, options.Container, []string{"/bin/sh"}, nil)
}

func (r Remoter) getTaskId(ctx context.Context, options admin.RemoteOptions) (string, error) {
	if options.Task == "" {
		if taskId, err := GetRandomTask(ctx, r.Infra); err != nil {
			return "", err
		} else if taskId == "" {
			return "", fmt.Errorf("cannot exec command with no running tasks")
		} else {
			return taskId, nil
		}
	}
	return options.Task, nil
}
