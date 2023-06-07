package k8s

import (
	"bufio"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"log"
	"os"
	"time"
)

var (
	logger               = log.New(os.Stderr, "", 0)
	infoLogger           = log.New(logger.Writer(), "    ", 0)
	bold                 = color.New(color.Bold)
	normal               = color.New()
	getPodTimeout        = 20 * time.Second
	maxFollowConcurrency = 10
)

type LogStreamer struct {
	OsWriters    logging.OsWriters
	Details      app.Details
	AppNamespace string
	AppName      string
	NewClientFn  func(ctx context.Context) (*kubernetes.Clientset, error)
}

func (l LogStreamer) Stream(ctx context.Context, options config.LogStreamOptions) error {
	appLabel := fmt.Sprintf("nullstone.io/app=%s", l.AppName)

	client, err := l.NewClientFn(ctx)
	if err != nil {
		return fmt.Errorf("error creating kubernetes client: %w", err)
	}
	pods, err := client.CoreV1().Pods(l.AppNamespace).List(ctx, metav1.ListOptions{LabelSelector: appLabel})
	if err != nil {
		return fmt.Errorf("error looking for app pods: %w", err)
	}
	if len(pods.Items) <= 0 {
		return fmt.Errorf("no pods found for app %q in namespace %q", l.AppName, l.AppNamespace)
	}

	// TODO: restClientGetter
	logOptions := NewPodLogOptions(options)
	requests, err := polymorphichelpers.LogsForObjectFn(nil, pods, logOptions, getPodTimeout, true)
	if err != nil {
		return err
	}

	if o.Follow && len(requests) > 1 {
		if len(requests) > maxFollowConcurrency {
			return fmt.Errorf(
				"you are attempting to follow %d log streams, but maximum allowed concurrency is %d, use --max-log-requests to increase the limit",
				len(requests), o.MaxFollowConcurrency,
			)
		}

		return l.followLogs(requests)
	}

	return l.emitSequential(ctx, requests)
}

func (l LogStreamer) followLogs(requests map[corev1.ObjectReference]rest.ResponseWrapper) error {
	stdout := l.OsWriters.Stdout()
}

func (l LogStreamer) emitSequential(ctx context.Context, requests map[corev1.ObjectReference]rest.ResponseWrapper) error {
	stdout := l.OsWriters.Stdout()
	for objRef, request := range requests {
		if err := l.writeRequest(ctx, objRef, request, stdout); err != nil {
			return err
		}
	}
	return nil
}

func (l LogStreamer) writeRequest(ctx context.Context, objRef corev1.ObjectReference, request rest.ResponseWrapper, out io.Writer) error {
	readCloser, err := request.Stream(ctx)
	if err != nil {
		return err
	}
	defer readCloser.Close()

	// TODO: Set prefix using pod name, container name
	prefix := fmt.Sprintf("")

	r := bufio.NewReader(readCloser)
	for {
		bytes, err := r.ReadBytes('\n')
		if _, err := bold.Fprintf(out, "[%s] ", prefix); err != nil {
			return err
		}
		if _, err := normal.Fprint(out, string(bytes)); err != nil {
			return err
		}
		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
	}
}
