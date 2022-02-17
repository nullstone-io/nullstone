package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func CancellableAction(fn func(ctx context.Context) error) error {
	ctx := context.Background()
	signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	return fn(ctx)
}
