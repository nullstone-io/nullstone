package cloudrun

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/run/apiv2/runpb"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/gcp"
	"github.com/nullstone-io/deployment-sdk/gcp/cloudrun"
	"golang.org/x/sync/errgroup"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
)

type JobRunner struct {
	JobId             string
	MainContainerName string
	Adminer           gcp.ServiceAccount
}

const (
	// Polling interval for checking job status
	defaultPollInterval = 2 * time.Second
)

func (r JobRunner) Run(ctx context.Context, options admin.RunOptions, cmd []string, envVars map[string]string) error {
	// Start the job execution
	exec, err := r.startJob(ctx, cmd, envVars)
	if err != nil {
		return fmt.Errorf("error running job: %w", err)
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		absoluteTime := time.Now()
		executionFilter := fmt.Sprintf("resource.labels.execution_id=%q", exec.Uid)
		logStreamOptions := app.LogStreamOptions{
			StartTime:     &absoluteTime,
			WatchInterval: time.Duration(0), // this makes sure the log stream doesn't exist until the context is canceled
			Emitter:       options.LogEmitter,
			Selector:      &executionFilter,
		}
		if err := options.LogStreamer.Stream(ctx, logStreamOptions); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
		return nil
	})

	// TODO: How do we report start failures that aren't reported to logs? (e.g. container pull error)

	// Monitor the job execution
	eg.Go(func() error {
		return r.monitorExecution(ctx, exec)
	})

	// Wait for all goroutines to complete
	return eg.Wait()
}

func (r JobRunner) startJob(ctx context.Context, cmd []string, envVars map[string]string) (*runpb.Execution, error) {
	// Create a client for Cloud Run Jobs
	client, err := cloudrun.NewJobsClient(ctx, r.Adminer)
	if err != nil {
		return nil, fmt.Errorf("error creating Cloud Run Jobs client: %w", err)
	}

	// If a command is provided, use it as "args" in main container overrides
	req := &runpb.RunJobRequest{
		Name:      r.JobId,
		Overrides: &runpb.RunJobRequest_Overrides{ContainerOverrides: make([]*runpb.RunJobRequest_Overrides_ContainerOverride, 0)},
	}

	var args []string
	var envOverrides []*runpb.EnvVar
	for name, value := range envVars {
		envOverrides = append(envOverrides, &runpb.EnvVar{Name: name, Values: &runpb.EnvVar_Value{Value: value}})
	}
	if len(cmd) > 0 {
		args = cmd
	}

	if len(cmd) > 0 || len(envVars) > 0 {
		req.Overrides = &runpb.RunJobRequest_Overrides{
			ContainerOverrides: []*runpb.RunJobRequest_Overrides_ContainerOverride{
				{
					Name: r.MainContainerName,
					Args: args,
					Env:  envOverrides,
				},
			},
		}
	}

	op, err := client.RunJob(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error starting job: %w", err)
	}
	if _, err := op.Wait(ctx); err != nil {
		return nil, fmt.Errorf("an error occurred waiting for job to start: %w", err)
	}
	return op.Metadata()
}

// monitorExecution monitors the job execution until it completes
func (r JobRunner) monitorExecution(ctx context.Context, exec *runpb.Execution) error {
	client, err := cloudrun.NewExecutionsClient(ctx, r.Adminer)
	if err != nil {
		return fmt.Errorf("error creating Cloud Run Executions client: %w", err)
	}

	ticker := time.NewTicker(defaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			updated, err := client.GetExecution(ctx, &runpb.GetExecutionRequest{Name: exec.Name})
			if err != nil {
				return fmt.Errorf("error getting job execution status: %w", err)
			}

			if (updated.FailedCount + updated.SucceededCount) >= updated.TaskCount {
				// Job execution completed
				if updated.FailedCount == 1 && updated.TaskCount == 1 {
					return fmt.Errorf("job execution failed")
				}
				if updated.FailedCount > 1 {
					return fmt.Errorf("%d of %d job executions failed", updated.FailedCount, updated.TaskCount)
				}
				return nil
			}
		}
	}
}
