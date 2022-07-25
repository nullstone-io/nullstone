package admin

import (
	"context"
	"gopkg.in/nullstone-io/nullstone.v0/config"
)

type LogStreamer interface {
	Stream(ctx context.Context, options config.LogStreamOptions) error
}
