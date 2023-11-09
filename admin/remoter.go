package admin

import (
	"context"
	"gopkg.in/nullstone-io/nullstone.v0/config"
)

type RemoteOptions struct {
	// Task refers to the ECS task id for remote access if using ECS
	Task string
	// Pod refers to the k8s pod for remote access if using k8s
	Pod string
	// Container represents the specific container name for remote access in the k8s pod or ecs task
	Container    string
	PortForwards []config.PortForward
	Username     string
	LogStreamer  LogStreamer
}

type Remoter interface {
	// Exec allows a user to execute a command (usually tunneling) into a running service
	// This only makes sense for container-based providers
	Exec(ctx context.Context, options RemoteOptions, cmd []string) error

	// Ssh allows a user to SSH into a running service
	Ssh(ctx context.Context, options RemoteOptions) error
}
