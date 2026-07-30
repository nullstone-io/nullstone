package admin

import (
	"context"

	"github.com/nullstone-io/deployment-sdk/app"
	"gopkg.in/nullstone-io/nullstone.v0/config"
)

type RemoteOptions struct {
	// Instance refers to the VM Instance for remote access
	Instance string
	// Task refers to the ECS task id for remote access if using ECS
	Task string
	// Pod refers to the k8s pod for remote access if using k8s
	Pod string
	// Container represents the specific container name for remote access in the k8s pod or ecs task
	Container    string
	PortForwards []config.PortForward
	Username     string
	LogStreamer  app.LogStreamer
	LogEmitter   app.LogEmitter
}

const (
	// TriggerEnvVar identifies how the job/task was started
	TriggerEnvVar = "NULLSTONE_TRIGGER"
	// TriggerNameEnvVar identifies who started the job/task
	TriggerNameEnvVar = "NULLSTONE_TRIGGER_NAME"

	// TriggerManual is the trigger for a job/task started through the CLI
	TriggerManual = "manual"
)

type RunOptions struct {
	// Container represents the specific container name to execute against in the k8s pod/ecs task
	Container string
	// Payload contains the input event for providers that invoke a function instead of running a command
	// This is only supported for AWS Lambda
	Payload     []byte
	Username    string
	LogStreamer app.LogStreamer
	LogEmitter  app.LogEmitter
}

type Remoter interface {
	// Exec allows a user to execute a command (usually tunneling) into a running service
	// This only makes sense for container-based providers
	Exec(ctx context.Context, options RemoteOptions, cmd []string) error

	// Ssh allows a user to SSH into a running service
	Ssh(ctx context.Context, options RemoteOptions) error

	// Run starts a new job/task and executes a command
	Run(ctx context.Context, options RunOptions, cmd []string, envVars map[string]string) error
}
