package mcp

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	// DefaultTimeout is the maximum time a command can run before being killed.
	// Infrastructure commands with --wait can take a long time to complete.
	DefaultTimeout = 30 * time.Minute
)

// ExecuteResult holds the captured output from a CLI command execution.
type ExecuteResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Execute runs a nullstone CLI command as a subprocess, capturing all output.
// It uses os.Args[0] to invoke the same binary, with stdout/stderr captured
// into buffers (completely separate from the MCP server's own stdio).
func Execute(ctx context.Context, args []string) (*ExecuteResult, error) {
	binaryPath := os.Args[0]

	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// Do not set cmd.Stdin — child processes must not read from the MCP server's stdin

	err := cmd.Run()

	result := &ExecuteResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return result, fmt.Errorf("failed to execute command: %w", err)
		}
	}

	return result, nil
}

// BuildArgs constructs the CLI argument list from the tool name and parameters.
// Global flags (--profile, --org) are placed before the subcommand.
// Parameter names with underscores are converted to hyphen-delimited flags.
func BuildArgs(toolName string, params map[string]interface{}) []string {
	var args []string

	// Global flags go first (before the subcommand)
	if profile, ok := getStringParam(params, "profile"); ok && profile != "" {
		args = append(args, "--profile", profile)
	}
	if org, ok := getStringParam(params, "org"); ok && org != "" {
		args = append(args, "--org", org)
	}

	// Convert tool name to subcommand(s)
	subcommands := toolNameToSubcommands(toolName)
	args = append(args, subcommands...)

	// Parameters to skip when building flags
	skipParams := map[string]bool{
		"profile": true,
		"org":     true,
	}
	positionalKeys := toolPositionalArgs(toolName)
	for _, key := range positionalKeys {
		skipParams[key] = true
	}

	// Add command-specific flags
	for key, value := range params {
		if skipParams[key] {
			continue
		}
		addFlagArgs(&args, key, value)
	}

	// Add positional args last
	for _, key := range positionalKeys {
		if val, ok := getStringParam(params, key); ok && val != "" {
			args = append(args, val)
		}
	}

	return args
}

// toolNameToSubcommands converts an MCP tool name to CLI subcommand arguments.
func toolNameToSubcommands(toolName string) []string {
	mapping := map[string][]string{
		"set_org":           {"set-org"},
		"stacks_list":       {"stacks", "list"},
		"stacks_new":        {"stacks", "new"},
		"envs_list":         {"envs", "list"},
		"envs_new":          {"envs", "new"},
		"envs_delete":       {"envs", "delete"},
		"envs_up":           {"envs", "up"},
		"envs_down":         {"envs", "down"},
		"apps_list":         {"apps", "list"},
		"blocks_list":       {"blocks", "list"},
		"blocks_new":        {"blocks", "new"},
		"workspaces_select": {"workspaces", "select"},
		"iac_test":          {"iac", "test"},
		"iac_generate":      {"iac", "generate"},
		"modules_register":  {"modules", "register"},
		"modules_publish":   {"modules", "publish"},
		"modules_package":   {"modules", "package"},
	}

	if cmds, ok := mapping[toolName]; ok {
		return cmds
	}

	// Top-level commands: outputs, status, profile, logs, up, plan, apply, wait, push, deploy, launch
	return []string{toolName}
}

// toolPositionalArgs returns parameter names that should be passed as positional
// arguments (not flags) for a given tool.
func toolPositionalArgs(toolName string) []string {
	switch toolName {
	case "set_org":
		return []string{"org_name"}
	default:
		return nil
	}
}

// addFlagArgs appends a parameter as CLI flag argument(s).
func addFlagArgs(args *[]string, key string, value interface{}) {
	flagName := "--" + strings.ReplaceAll(key, "_", "-")

	switch v := value.(type) {
	case bool:
		if v {
			*args = append(*args, flagName)
		}
	case string:
		if v != "" {
			*args = append(*args, flagName, v)
		}
	case float64:
		*args = append(*args, flagName, fmt.Sprintf("%v", v))
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				*args = append(*args, flagName, s)
			}
		}
	}
}

func getStringParam(params map[string]interface{}, key string) (string, bool) {
	if v, ok := params[key]; ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}
