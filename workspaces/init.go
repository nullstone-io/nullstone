package workspaces

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Init(ctx context.Context, toolName string) error {
	var process string
	switch toolName {
	default:
		process = "terraform"
	case "terraform":
		process = "terraform"
	case "opentofu":
		process = "tofu"
	}

	args := []string{
		"init",
		"-reconfigure",
	}
	fmt.Printf("Running `%s %s`\n", process, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, process, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
