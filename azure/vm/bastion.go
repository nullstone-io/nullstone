package vm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

const (
	azCliBinary = "az"
	azCliUrl    = "https://learn.microsoft.com/en-us/cli/azure/install-azure-cli"
)

// StartBastionSsh initiates an interactive SSH session with an Azure VM via Azure Bastion.
// This shells out to `az network bastion ssh` which handles the Bastion tunnel,
// mirroring the pattern used by AWS SSM (session-manager-plugin subprocess).
func StartBastionSsh(ctx context.Context, infra Outputs, username string) error {
	azPath, err := getAzCliPath()
	if err != nil {
		return err
	}

	token, err := infra.Remoter.GetToken(ctx, policy.TokenRequestOptions{})
	if err != nil {
		return fmt.Errorf("error getting Azure token: %w", err)
	}

	vmResourceId := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s",
		infra.SubscriptionId, infra.ResourceGroup, infra.VmName)

	args := []string{
		"network", "bastion", "ssh",
		"--name", infra.BastionHostName,
		"--resource-group", infra.ResourceGroup,
		"--target-resource-id", vmResourceId,
		"--auth-type", "AAD",
	}
	if username != "" {
		args = append(args, "--username", username)
	}

	// Set the access token so the az CLI uses it instead of requiring a separate login
	env := os.Environ()
	env = append(env, "AZURE_ACCESS_TOKEN="+token.Token)

	ctx = context.Background() // Ignore signal cancellations on the context
	cmd := exec.CommandContext(ctx, azPath, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Env = env

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting bastion ssh session: %w", err)
	}

	done := make(chan any)
	defer close(done)
	forwardSignals(done, cmd.Process)
	return cmd.Wait()
}

func forwardSignals(done chan any, process *os.Process) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch)
	go func() {
		select {
		case <-done:
			return
		case sig := <-ch:
			process.Signal(sig)
		}
	}()
}

// getAzCliPath attempts to find the "az" CLI binary.
func getAzCliPath() (string, error) {
	if _, err := exec.LookPath(azCliBinary); err == nil {
		return azCliBinary, nil
	}
	return "", fmt.Errorf("could not find Azure CLI (az). Visit %q to install", azCliUrl)
}
