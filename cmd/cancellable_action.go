package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func CancellableAction(fn func(ctx context.Context) error) error {
	ctx := context.Background()
	// Handle Ctrl+C, kill stream
	ctx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-term
		cancelFn()
	}()
	return fn(ctx)
}
