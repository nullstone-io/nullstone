package app_logs

import (
	"context"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/config"
)

type Provider interface {
	Stream(ctx context.Context, nsConfig api.Config, app *types.Application, workspace *types.Workspace, options config.LogStreamOptions) error
}
