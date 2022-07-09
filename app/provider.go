package app

import "context"

type Provider interface {
	NewPusher() (Pusher, error)
	NewDeployer() (Deployer, error)
	NewDeployStatusGetter() (DeployStatusGetter, error)
}

type Pusher interface {
	Push(ctx context.Context, source, version string) error
}

type Deployer interface {
	Deploy(ctx context.Context, version string) (*string, error)
}

type DeployStatusGetter interface {
	GetDeployStatus(ctx context.Context, reference *string) error
}
