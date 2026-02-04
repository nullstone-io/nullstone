package workspaces

import (
	"context"
	"os"
	"os/exec"
)

func Init(ctx context.Context, toolName string) error {
	var process string
	switch toolName {
	default:
		process = "terraform"
	case "terraform":
		process = "terraform"
	case "opentofu":
		process = "opentofu"
	}

	args := []string{
		"init",
		"-reconfigure",
	}
	cmd := exec.CommandContext(ctx, process, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
