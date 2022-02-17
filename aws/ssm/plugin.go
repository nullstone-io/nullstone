package ssm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"os"
	"os/exec"
)

const (
	sessionManagerBinary = "session-manager-plugin"
	sessionManagerUrl    = "https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html"
)

func StartSession(ctx context.Context, session interface{}, target ssm.StartSessionInput, region, endpointUrl string) error {
	process, err := getSessionManagerPluginPath()
	if err != nil {
		return fmt.Errorf("could not find AWS session-manager-plugin: %w", err)
	}

	sessionJsonRaw, _ := json.Marshal(session)
	targetRaw, _ := json.Marshal(target)
	args := []string{
		string(sessionJsonRaw),
		region,
		"StartSession",
		"", // empty profile name
		string(targetRaw),
		endpointUrl,
	}
	cmd := exec.CommandContext(ctx, process, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// getSessionManagerPluginPath attempts to find "session-manager-plugin"
// If it's in PATH, will simply return binary name
// If not, will attempt OS-specific locations
func getSessionManagerPluginPath() (string, error) {
	if _, err := exec.LookPath(sessionManagerBinary); err == nil {
		return sessionManagerBinary, nil
	}
	if _, err := os.Stat(osSessionManagerPluginPath); err != nil {
		return "", fmt.Errorf("Could not find session-manager-plugin. Visit %q to install.", sessionManagerUrl)
	}
	return osSessionManagerPluginPath, nil
}
