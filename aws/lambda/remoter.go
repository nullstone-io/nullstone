package lambda

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
)

var (
	_ admin.Remoter = Remoter{}
)

func NewRemoter(ctx context.Context, osWriters logging.OsWriters, source outputs.RetrieverSource, appDetails app.Details) (admin.Remoter, error) {
	outs, err := outputs.Retrieve[Outputs](ctx, source, appDetails.Workspace, appDetails.WorkspaceConfig)
	if err != nil {
		return nil, err
	}
	outs.InitializeCreds(source, appDetails.Workspace)

	return Remoter{
		OsWriters: osWriters,
		Details:   appDetails,
		Infra:     outs,
	}, nil
}

type Remoter struct {
	OsWriters logging.OsWriters
	Details   app.Details
	Infra     Outputs
}

func (r Remoter) Exec(ctx context.Context, options admin.RemoteOptions, cmd []string) error {
	return fmt.Errorf("cannot `exec` into a lambda function; use `run` if you want to invoke the function")
}

func (r Remoter) Ssh(ctx context.Context, options admin.RemoteOptions) error {
	return fmt.Errorf("cannot `ssh` into a lambda function")
}

func (r Remoter) Run(ctx context.Context, options admin.RunOptions, cmd []string, envVars map[string]string) error {
	// A lambda function has no entrypoint to override; it receives an input event instead of a command
	if len(cmd) > 0 {
		return fmt.Errorf("a lambda function does not run a command; use --payload to send an input event to the function")
	}
	// Lambda environment variables are baked into the function configuration at deploy time
	// The Invoke API has no mechanism to override them for a single invocation
	if names := userEnvVarNames(envVars); len(names) > 0 {
		return fmt.Errorf("cannot set environment variables (%s) when running a lambda function; "+
			"lambda environment variables are configured at deploy time, use --payload to send input to the function",
			strings.Join(names, ", "))
	}
	return Invoke(ctx, r.OsWriters, r.Infra, options)
}

// userEnvVarNames filters out the env vars that the CLI attaches to every run
// Anything left over was specified by the user through --env-var
func userEnvVarNames(envVars map[string]string) []string {
	names := make([]string, 0)
	for name := range envVars {
		if _, ok := triggerEnvVars[name]; ok {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
