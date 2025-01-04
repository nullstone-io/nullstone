package k8s

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/k8s"
	"golang.org/x/sync/errgroup"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"strings"
	"time"
)

const DefaultWatchInterval = 1 * time.Second

type JobRunner struct {
	Namespace         string
	MainContainerName string
	JobDefinition     string
	NewConfigFn       k8s.NewConfiger
}

func (r JobRunner) Run(ctx context.Context, options admin.RunOptions, cmd []string) error {
	decoder := json.NewDecoder(base64.NewDecoder(base64.StdEncoding, strings.NewReader(r.JobDefinition)))
	var jobDef batchv1.Job
	if err := decoder.Decode(&jobDef); err != nil {
		return fmt.Errorf("job_definition from app module outputs is invalid: %w", err)
	}

	cfg, err := r.NewConfigFn(ctx)
	if err != nil {
		return fmt.Errorf("error creating kubernetes config: %w", err)
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error creating kube client: %w", err)
	}

	// Add a unique suffix (-{{timestamp}}) to ensure we can `run` repeatedly to create new jobs
	jobDef.Name = fmt.Sprintf("%s-%d", jobDef.Name, time.Now().Unix())

	// Override `command` if specified by CLI user
	r.overrideMainContainerCommand(jobDef, cmd)

	job, err := client.BatchV1().Jobs(r.Namespace).Create(ctx, &jobDef, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("error creating job: %w", err)
	}

	return r.monitorJob(ctx, client, options, job.Name)
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
		return options.LogStreamer.Stream(ctx, logStreamOptions)
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
					if exitCode := containerStatus.State.Terminated.ExitCode; exitCode == 0 {
						log.Printf("Job has completed successfully")
						return nil
					} else {

						return fmt.Errorf("Job failed with status code %d\n", exitCode)
					}
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

func (r JobRunner) overrideMainContainerCommand(job batchv1.Job, cmd []string) batchv1.Job {
	if len(cmd) < 1 {
		return job
	}
	for i, container := range job.Spec.Template.Spec.Containers {
		if container.Name == r.MainContainerName {
			container.Command = cmd
			job.Spec.Template.Spec.Containers[i] = container
			return job
		}
	}
	return job
}
