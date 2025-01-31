package k8s

import (
	"fmt"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"net/http"
	"net/url"
)

func ForwardPorts(stop <-chan struct{}, transport http.RoundTripper, upgrader spdy.Upgrader, opts *ExecOptions, url *url.URL) {
	if len(opts.PortMappings) < 1 {
		return
	}

	ready := make(chan struct{})
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, url)
	fw, err := portforward.New(dialer, opts.PortMappings, stop, ready, opts.Out, opts.ErrOut)
	if err != nil {
		fmt.Fprintf(opts.ErrOut, "error forwarding ports: %s\n", err)
		return
	}

	if err := fw.ForwardPorts(); err != nil {
		fmt.Fprintf(opts.ErrOut, "port forwarding stopped: %s\n", err)
	}
}
