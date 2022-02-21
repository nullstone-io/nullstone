package workspaces

import (
	"context"
	"os"
	"os/exec"
)

func Init(ctx context.Context) error {
	process := "terraform"
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
