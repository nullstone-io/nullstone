package runs

import "gopkg.in/nullstone-io/go-api-client.v0/types"

// fillRunConfig ensures that values are configured with their defaults
func fillRunConfig(rc *types.RunConfig) {
	rc.Variables = fillRunConfigVariables(rc.Variables)
	for i, c := range rc.Capabilities {
		c.Variables = fillRunConfigVariables(c.Variables)
		rc.Capabilities[i] = c
	}
	for i, dc := range rc.DependencyConfigs {
		dc.Variables = fillRunConfigVariables(dc.Variables)
		rc.DependencyConfigs[i] = dc
	}
}

func fillRunConfigVariables(vars types.Variables) types.Variables {
	for k, v := range vars {
		if v.Value == nil {
			v.Value = v.Default
		}
		vars[k] = v
	}
	return vars
}
