package cmd

import (
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/urfave/cli/v2"
	"io"
)

var _ logging.OsWriters = CliOsWriters{}

type CliOsWriters struct {
	Context *cli.Context
}

func (c CliOsWriters) Stdout() io.Writer {
	return c.Context.App.Writer
}

func (c CliOsWriters) Stderr() io.Writer {
	return c.Context.App.ErrWriter
}
