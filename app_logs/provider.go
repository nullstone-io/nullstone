package app_logs

import (
	"context"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/nullstone.v0/app"
	"gopkg.in/nullstone-io/nullstone.v0/config"
)

type Provider interface {
	Stream(ctx context.Context, nsConfig api.Config, details app.Details, options config.LogStreamOptions) error
}
