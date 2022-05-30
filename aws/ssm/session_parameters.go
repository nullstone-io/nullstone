package ssm

import (
	"fmt"
	"gopkg.in/nullstone-io/nullstone.v0/config"
)

func SessionParametersFromPortForwards(pfs []config.PortForward) (map[string][]string, error) {
	switch len(pfs) {
	case 0:
		return nil, nil
	case 1:
		m := map[string][]string{
			"localPort":  {pfs[0].LocalPort},
			"portNumber": {pfs[0].RemotePort},
		}
		if pfs[0].RemoteHost != "" {
			m["host"] = []string{pfs[0].RemoteHost}
		}
		return m, nil
	default:
		return nil, fmt.Errorf("AWS does not support more than one port forward.")
	}
}
