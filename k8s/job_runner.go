package k8s

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/k8s"
	"golang.org/x/sync/errgroup"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const DefaultWatchInterval = 1 * time.Second

type JobRunner struct {
	Namespace         string
	AppName           string
	MainContainerName string
	JobDefinitionName string
	NewConfigFn       k8s.NewConfiger

	Out    io.Writer
	ErrOut io.Writer
}

func (r JobRunner) Run(ctx context.Context, options admin.RunOptions, cmd []string) error {
	cfg, err := r.NewConfigFn(ctx)
	if err != nil {
		return fmt.Errorf("error creating kubernetes config: %w", err)
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error creating kube client: %w", err)
	}

	fmt.Fprintln(r.ErrOut, "Retrieving job definition...")
	jobDef, _, err := k8s.GetJobDefinition(ctx, client, r.Namespace, r.JobDefinitionName)
	if err != nil {
		return fmt.Errorf("error retrieving job definition from Kubernetes config map (%s): %w", r.JobDefinitionName, err)
	}

	// Create a unique job name (e.g. `<app-name>-<timestamp>`) so we can repeatedly create new jobs
	jobDef.Name = fmt.Sprintf("%s-%d", r.AppName, time.Now().Unix())
	// If specified by a CLI user, override cmd in the "main" container
	jobDef.Spec.Template = r.overrideMainContainerCommand(jobDef.Spec.Template, cmd)

	fmt.Fprintf(r.ErrOut, "Creating kubernetes job (name = %s)...\n", jobDef.Name)
	job, err := client.BatchV1().Jobs(r.Namespace).Create(ctx, jobDef, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("error creating job: %w", err)
	}

	fmt.Fprintln(r.ErrOut, "Waiting for job to start...")
	if err := r.waitForActiveJob(ctx, client, jobDef.Name); err != nil {
		return err
	}
	fmt.Fprintln(r.ErrOut, "Job started.")
	
	fmt.Fprintln(r.ErrOut, "Monitoring job and streaming logs...")
	return r.monitorJob(ctx, client, options, job.Name)
}

func (r JobRunner) waitForActiveJob(ctx context.Context, client *kubernetes.Clientset, jobName string) error {
	for {
		job, err := client.BatchV1().Jobs(r.Namespace).Get(ctx, jobName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting status of job: %w", err)
		}
		if job.Status.Failed > 0 {
			return fmt.Errorf("Job failed to start")
		}
		if job.Status.Active > 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			switch err := ctx.Err(); {
			case errors.Is(err, context.Canceled):
				return fmt.Errorf("cancelled")
			case errors.Is(err, context.DeadlineExceeded):
				return fmt.Errorf("timeout")
			}
		case <-time.After(DefaultWatchInterval):
		}
	}
}

func (r JobRunner) monitorJob(ctx context.Context, client *kubernetes.Clientset, options admin.RunOptions, jobName string) error {
	eg, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)

	selector := fmt.Sprintf("job-name=%s", jobName)

	eg.Go(func() error {
		absoluteTime := time.Now()
		logStreamOptions := app.LogStreamOptions{
			StartTime:     &absoluteTime,
			WatchInterval: time.Duration(0), // this makes sure the log stream doesn't exit until the context is cancelled
			Emitter:       options.LogEmitter,
			Selector:      &selector,
		}
		if err := options.LogStreamer.Stream(ctx, logStreamOptions); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
		return nil
	})
	eg.Go(func() error {
		defer cancel()
		for {
			// check status of job
			containerStatus, err := r.getJobContainerStatus(ctx, client, jobName)
			if err != nil {
				return err
			}
			if containerStatus != nil {
				if containerStatus.State.Terminated != nil {
					if exitCode := containerStatus.State.Terminated.ExitCode; exitCode != 0 {
						return fmt.Errorf("Job failed with status code %d\n", exitCode)
					}
					time.Sleep(time.Second) // Wait for logs to flush
					fmt.Printf("Job has completed successfully\n")
					return nil
				}
			}

			select {
			case <-ctx.Done():
				switch err := ctx.Err(); {
				case errors.Is(err, context.Canceled):
					return fmt.Errorf("cancelled")
				case errors.Is(err, context.DeadlineExceeded):
					return fmt.Errorf("timeout")
				}
			case <-time.After(DefaultWatchInterval):
			}
		}
	})
	return eg.Wait()
}

func (r JobRunner) getJobContainerStatus(ctx context.Context, client *kubernetes.Clientset, jobName string) (*corev1.ContainerStatus, error) {
	podsResponse, err := client.CoreV1().Pods(r.Namespace).List(ctx, metav1.ListOptions{LabelSelector: fmt.Sprintf("job-name=%s", jobName)})
	if err != nil {
		return nil, err
	}
	for _, pod := range podsResponse.Items {
		for _, cstatus := range pod.Status.ContainerStatuses {
			if cstatus.Name == r.MainContainerName || r.MainContainerName == "" {
				return &cstatus, nil
			}
		}
	}
	return nil, nil
}

func (r JobRunner) overrideMainContainerCommand(podTemplateSpec corev1.PodTemplateSpec, cmd []string) corev1.PodTemplateSpec {
	if len(cmd) < 1 {
		return podTemplateSpec
	}
	for i, container := range podTemplateSpec.Spec.Containers {
		if container.Name == r.MainContainerName {
			container.Command = cmd
			podTemplateSpec.Spec.Containers[i] = container
			return podTemplateSpec
		}
	}
	return podTemplateSpec
}
