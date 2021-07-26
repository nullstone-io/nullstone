package cmd

import (
	"bytes"
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
		buf := bytes.NewBufferString("")
		if err := fn(buf); err != nil {
			return err
		}
		io.Copy(writer, buf)
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
