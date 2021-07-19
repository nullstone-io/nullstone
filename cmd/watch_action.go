package cmd

import (
	"context"
	"github.com/gosuri/uilive"
	"io"
	"time"
)

func WatchAction(ctx context.Context, watchInterval time.Duration, fn func(w io.Writer) error) error {
	writer := uilive.New()
	writer.Start()
	defer writer.Stop()
	for {
		if err := fn(writer); err != nil {
			return err
		}
		if watchInterval <= 0*time.Second {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(watchInterval):
		}
	}
}
