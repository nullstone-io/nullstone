package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Serve creates and starts the MCP server over stdio transport.
func Serve() error {
	s := server.NewMCPServer(
		"nullstone",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	for _, tool := range AllTools() {
		s.AddTool(tool, makeHandler(tool.Name))
	}

	return server.ServeStdio(s)
}

// makeHandler creates an MCP tool handler that executes the corresponding
// CLI command via subprocess and returns the captured output.
func makeHandler(toolName string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := request.GetArguments()

		// envs_delete always gets --force since we can't prompt interactively
		if toolName == "envs_delete" {
			if params == nil {
				params = map[string]interface{}{}
			}
			params["force"] = true
		}

		args := BuildArgs(toolName, params)

		result, err := Execute(ctx, args)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to execute command: %s", err)), nil
		}

		var output strings.Builder
		if result.Stdout != "" {
			output.WriteString(result.Stdout)
		}
		if result.Stderr != "" {
			if output.Len() > 0 {
				output.WriteString("\n")
			}
			output.WriteString(result.Stderr)
		}

		if result.ExitCode != 0 {
			errMsg := output.String()
			if errMsg == "" {
				errMsg = fmt.Sprintf("Command exited with code %d", result.ExitCode)
			}
			return mcp.NewToolResultError(errMsg), nil
		}

		text := output.String()
		if text == "" {
			text = "Command completed successfully."
		}
		return mcp.NewToolResultText(text), nil
	}
}
