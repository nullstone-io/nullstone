package k8s

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"net/http"
	"os"
)

type PortForwarder struct {
	Transport http.RoundTripper
	Upgrader  spdy.Upgrader
	Request   *rest.Request
}

func NewPortForwarder(cfg *rest.Config, podNamespace, podName string, portMappings []string) (*PortForwarder, error) {
	if len(portMappings) < 1 {
		return nil, nil
	}

	restClient, err := rest.RESTClientFor(cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot create rest client: %w", err)
	}
	req := restClient.Post().
		Resource("pods").
		Namespace(podNamespace).
		Name(podName).
		SubResource("portforward").
		VersionedParams(&corev1.PodPortForwardOptions{}, scheme.ParameterCodec)

	transport, upgrader, err := spdy.RoundTripperFor(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create SPDY transport: %w", err)
	}

	return &PortForwarder{
		Transport: transport,
		Upgrader:  upgrader,
		Request:   req,
	}, nil
}

func (f *PortForwarder) ForwardPorts(stop <-chan struct{}, opts *ExecOptions) {
	if len(opts.PortMappings) < 1 {
		return
	}

	stderr := opts.ErrOut
	if stderr == nil {
		stderr = os.Stderr
	}

	ready := make(chan struct{})
	dialer := spdy.NewDialer(f.Upgrader, &http.Client{Transport: f.Transport}, http.MethodPost, f.Request.URL())
	fw, err := portforward.New(dialer, opts.PortMappings, stop, ready, opts.Out, opts.ErrOut)
	if err != nil {
		fmt.Fprintf(stderr, "error forwarding ports: %s\n", err)
		return
	}

	if err := fw.ForwardPorts(); err != nil {
		fmt.Fprintf(stderr, "port forwarding stopped: %s\n", err)
	}
}
