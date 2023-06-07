package k8s

import (
	"bufio"
	"context"
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"gopkg.in/nullstone-io/nullstone.v0/config"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"regexp"
	"sync"
	"time"
)

var (
	getPodTimeout                  = 20 * time.Second
	maxFollowConcurrency           = 10
	containerNameFromRefSpecRegexp = regexp.MustCompile(`spec\.(?:initContainers|containers|ephemeralContainers){(.+)}`)
)

type NewConfiger func(ctx context.Context) (*rest.Config, error)
type MessageEmitter func(w io.Writer, podName, containerName, line string) error

type LogStreamer struct {
	OsWriters    logging.OsWriters
	Details      app.Details
	AppNamespace string
	AppName      string
	NewConfigFn  NewConfiger
	Emitter      MessageEmitter
}

func (l LogStreamer) Stream(ctx context.Context, options config.LogStreamOptions) error {
	appLabel := fmt.Sprintf("nullstone.io/app=%s", l.AppName)

	cfg, err := l.NewConfigFn(ctx)
	if err != nil {
		return fmt.Errorf("error configuring kubernetes client: %w", err)
	}
	client, err := kubernetes.NewForConfig(cfg)
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

	logOptions := NewPodLogOptions(options)
	requests, err := polymorphichelpers.LogsForObjectFn(RestClientGetter{Config: cfg}, pods, logOptions, getPodTimeout, true)
	if err != nil {
		return err
	}

	if logOptions.Follow && len(requests) > 1 {
		if len(requests) > maxFollowConcurrency {
			tmpl := `You are attempting to follow %d log streams, exceeding the maximum allowed of %d. 
Restricting log streams to the first %d log streams.`
			fmt.Fprintf(l.OsWriters.Stderr(), tmpl,
				len(requests), maxFollowConcurrency, maxFollowConcurrency)
			newRequests := map[corev1.ObjectReference]rest.ResponseWrapper{}
			count := 0
			for k, v := range requests {
				newRequests[k] = v
				count++
				if count >= len(requests) {
					break
				}
			}
			requests = newRequests
		}
		return l.emitParallel(ctx, requests)
	}
	return l.emitSequential(ctx, requests)
}

func (l LogStreamer) emitParallel(ctx context.Context, requests map[corev1.ObjectReference]rest.ResponseWrapper) error {
	stdout, stderr := l.OsWriters.Stdout(), l.OsWriters.Stderr()

	reader, writer := io.Pipe()
	wg := &sync.WaitGroup{}
	wg.Add(len(requests))
	for ref, request := range requests {
		go func(ref corev1.ObjectReference, request rest.ResponseWrapper) {
			defer wg.Done()
			if err := l.writeRequest(ctx, ref, request); err != nil {
				fmt.Fprintf(stderr, "unable to write logs: %s\n", err)
			}
		}(ref, request)
	}

	go func() {
		wg.Wait()
		writer.Close()
	}()

	_, err := io.Copy(stdout, reader)
	return err
}

func (l LogStreamer) emitSequential(ctx context.Context, requests map[corev1.ObjectReference]rest.ResponseWrapper) error {
	for ref, request := range requests {
		if err := l.writeRequest(ctx, ref, request); err != nil {
			return err
		}
	}
	return nil
}

func (l LogStreamer) writeRequest(ctx context.Context, ref corev1.ObjectReference, request rest.ResponseWrapper) error {
	stdout := l.OsWriters.Stdout()
	readCloser, err := request.Stream(ctx)
	if err != nil {
		return err
	}
	defer readCloser.Close()

	podName, containerName := l.parseRef(ref)

	r := bufio.NewReader(readCloser)
	for {
		str, readErr := r.ReadString('\n')
		if err := l.Emitter(stdout, podName, containerName, str); err != nil {
			return err
		}
		if readErr != nil {
			if readErr == io.EOF {
				return nil
			}
			return readErr
		}
	}
}

func (l LogStreamer) parseRef(ref corev1.ObjectReference) (string, string) {
	if ref.FieldPath == "" || ref.Name == "" {
		return ref.Name, ""
	}

	// We rely on ref.FieldPath to contain a reference to a container
	// including a container name (not an index) so we can get a container name
	// without making an extra API request.
	var containerName string
	containerNameMatches := containerNameFromRefSpecRegexp.FindStringSubmatch(ref.FieldPath)
	if len(containerNameMatches) == 2 {
		containerName = containerNameMatches[1]
	}

	return ref.Name, containerName
}
