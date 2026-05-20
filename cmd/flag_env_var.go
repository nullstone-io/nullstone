package cmd

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

// EnvVarFlag allows setting additional environment variables during a deploy.
// It can be specified multiple times, e.g. `--env-var FOO=bar --env-var BAZ=qux`.
var EnvVarFlag = &cli.StringSliceFlag{
	Name:    "env-var",
	Aliases: []string{"e"},
	Usage: `Set an additional environment variable on the app for this deployment in the form KEY=VALUE.
		Can be specified multiple times. Values may reference other --env-var values or standard env vars
		(e.g. NULLSTONE_VERSION) using {{ VAR }}, but cannot reference secrets.
		These env vars apply to this deployment only; a subsequent infra run or deploy may overwrite them.`,
}

// ParseEnvVars parses the repeated --env-var KEY=VALUE flag into a map.
func ParseEnvVars(c *cli.Context) (map[string]string, error) {
	raw := c.StringSlice(EnvVarFlag.Name)
	if len(raw) == 0 {
		return nil, nil
	}

	envVars := make(map[string]string, len(raw))
	for _, kvp := range raw {
		tokens := strings.SplitN(kvp, "=", 2)
		if len(tokens) < 2 || tokens[0] == "" {
			return nil, fmt.Errorf("invalid --env-var %q: must be in the form KEY=VALUE", kvp)
		}
		envVars[tokens[0]] = tokens[1]
	}
	return envVars, nil
}
