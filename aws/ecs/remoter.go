package ecs

import (
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
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
	if len(options.PortForwards) > 0 {
		return fmt.Errorf("ecs provider does not support port forwarding")
	}
	if r.Infra.ServiceName == "" {
		if options.Task != "" {
			return fmt.Errorf("fargate and ecs tasks do not support selecting a task, this exec command starts a new task")
		}
		return RunTask(ctx, r.Infra, options.Container, options.Username, cmd, options.LogStreamer, options.LogEmitter)
	}
	taskId, err := r.getTaskId(ctx, options)
	if err != nil {
		return err
	}
	return ExecCommand(ctx, r.Infra, taskId, options.Container, cmd, nil)
}

func (r Remoter) Ssh(ctx context.Context, options admin.RemoteOptions) error {
	if r.Infra.ServiceName == "" {
		return fmt.Errorf("fargate and ecs tasks do not support ssh")
	}
	if len(options.PortForwards) > 0 {
		return fmt.Errorf("ecs provider does not support port forwarding")
	}
	taskId, err := r.getTaskId(ctx, options)
	if err != nil {
		return err
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
