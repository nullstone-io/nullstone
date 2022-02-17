package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func CancellableAction(fn func(ctx context.Context) error) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return fn(ctx)
}
