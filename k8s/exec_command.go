package k8s

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/transport/spdy"
	"net/http"
)

func ExecCommand(ctx context.Context, cfg *rest.Config, podNamespace, podName, containerName string, cmd []string, opts *ExecOptions) error {
	tty, sizeQueue, err := opts.CreateTTY()
	if err != nil {
		return fmt.Errorf("unable to execute kubernetes command: %w", err)
	}

	return tty.Safe(func() error {
		restClient, err := rest.RESTClientFor(cfg)
		if err != nil {
			return err
		}

		req := restClient.Post().
			Resource("pods").
			Name(podName).
			Namespace(podNamespace).
			SubResource("exec")
		req.VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   cmd,
			Stdin:     opts.In != nil,
			Stdout:    opts.Out != nil,
			Stderr:    opts.ErrOut != nil,
			TTY:       tty.Raw,
		}, scheme.ParameterCodec)

		transport, upgrader, err := spdy.RoundTripperFor(cfg)
		if err != nil {
			return fmt.Errorf("unable to create kubernetes round tripper: %w", err)
		}
		executor, err := remotecommand.NewSPDYExecutorForTransports(transport, upgrader, http.MethodPost, req.URL())
		if err != nil {
			return fmt.Errorf("unable to create kubernetes remote executor: %w", err)
		}

		stop := make(chan struct{}, 1)
		defer close(stop)
		go ForwardPorts(stop, transport, upgrader, opts, req.URL())

		return executor.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdin:             opts.In,
			Stdout:            opts.Out,
			Stderr:            opts.ErrOut,
			Tty:               opts.TTY,
			TerminalSizeQueue: sizeQueue,
		})
	})
}
