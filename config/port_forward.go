package config

import (
	"fmt"
	"strings"
)

func ParsePortForward(arg string) (PortForward, error) {
	tokens := strings.SplitN(arg, ":", 3)
	switch len(tokens) {
	case 2:
		return PortForward{
			LocalPort:  tokens[0],
			RemoteHost: "",
			RemotePort: tokens[1],
		}, nil
	case 3:
		return PortForward{
			LocalPort:  tokens[0],
			RemoteHost: tokens[1],
			RemotePort: tokens[2],
		}, nil
	default:
		return PortForward{}, fmt.Errorf("expected <local-port>:[<remote-host>]:<remote-port>")
	}
}

type PortForward struct {
	LocalPort  string
	RemoteHost string
	RemotePort string
}
