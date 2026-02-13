package cmd

import (
	"github.com/urfave/cli/v2"
	"gopkg.in/nullstone-io/nullstone.v0/mcp"
)

var McpServer = &cli.Command{
	Name:        "mcp-server",
	Description: "Start an MCP (Model Context Protocol) server that exposes Nullstone CLI commands as tools for AI assistants. Communicates over stdio.",
	Usage:       "Start MCP server",
	UsageText:   "nullstone mcp-server",
	Action: func(c *cli.Context) error {
		return mcp.Serve()
	},
}
