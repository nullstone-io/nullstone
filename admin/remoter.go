package admin

import (
	"context"
	"gopkg.in/nullstone-io/nullstone.v0/config"
)

type Remoter interface {
	// Exec allows a user to execute a command (usually tunneling) into a running service
	// This only makes sense for container-based providers
	Exec(ctx context.Context, task string, cmd string) error

	// Ssh allows a user to SSH into a running service
	Ssh(ctx context.Context, task string, forwards []config.PortForward) error
}
