package cmd

import (
	"github.com/urfave/cli/v2"
)

var Workflows = &cli.Command{
	Name:      "workflows",
	Usage:     "View and cancel workflows (infrastructure runs, builds, and deploys) for a block in an environment",
	UsageText: "nullstone workflows [subcommand]",
	Subcommands: []*cli.Command{
		WorkflowsList,
		WorkflowsCancel,
	},
}
