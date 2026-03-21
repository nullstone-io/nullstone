package vm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

const armBaseURL = "https://management.azure.com"

// RunCommand executes a command on an Azure VM using the Run Command API
// POST /subscriptions/{subscriptionId}/resourceGroups/{resourceGroup}/providers/Microsoft.Compute/virtualMachines/{vmName}/runCommand
func RunCommand(ctx context.Context, infra Outputs, cmd []string) error {
	token, err := infra.Remoter.GetToken(ctx, policy.TokenRequestOptions{})
	if err != nil {
		return fmt.Errorf("error getting Azure token: %w", err)
	}

	// Azure Run Command expects a script as an array of strings (lines)
	script := []string{strings.Join(cmd, " ")}

	body := map[string]interface{}{
		"commandId": "RunShellScript",
		"script":    script,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshaling run command request: %w", err)
	}

	url := fmt.Sprintf("%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s/runCommand?api-version=2024-07-01",
		armBaseURL, infra.ResourceGroup, infra.VmName)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return fmt.Errorf("error creating run command request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error running command: %w", err)
	}
	defer resp.Body.Close()

	// Run Command returns 202 Accepted with an Azure-AsyncOperation header for polling
	if resp.StatusCode == http.StatusAccepted {
		asyncURL := resp.Header.Get("Azure-AsyncOperation")
		if asyncURL == "" {
			asyncURL = resp.Header.Get("Location")
		}
		if asyncURL != "" {
			return pollAsyncOperation(ctx, infra, asyncURL)
		}
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error running command (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result runCommandResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("error decoding run command result: %w", err)
	}

	for _, msg := range result.Value {
		if msg.Code == "ComponentStatus/StdOut/succeeded" {
			fmt.Print(msg.Message)
		} else if msg.Code == "ComponentStatus/StdErr/succeeded" && msg.Message != "" {
			fmt.Print(msg.Message)
		}
	}

	return nil
}

type runCommandResult struct {
	Value []runCommandMessage `json:"value"`
}

type runCommandMessage struct {
	Code    string `json:"code"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

func pollAsyncOperation(ctx context.Context, infra Outputs, asyncURL string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}

		token, err := infra.Remoter.GetToken(ctx, policy.TokenRequestOptions{})
		if err != nil {
			return fmt.Errorf("error getting Azure token: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, asyncURL, nil)
		if err != nil {
			return fmt.Errorf("error creating poll request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error polling operation: %w", err)
		}

		var status struct {
			Status     string            `json:"status"`
			Properties *runCommandResult `json:"properties"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			resp.Body.Close()
			return fmt.Errorf("error decoding poll response: %w", err)
		}
		resp.Body.Close()

		switch status.Status {
		case "Succeeded":
			if status.Properties != nil {
				for _, msg := range status.Properties.Value {
					if msg.Code == "ComponentStatus/StdOut/succeeded" {
						fmt.Print(msg.Message)
					} else if msg.Code == "ComponentStatus/StdErr/succeeded" && msg.Message != "" {
						fmt.Print(msg.Message)
					}
				}
			}
			return nil
		case "Failed":
			return fmt.Errorf("run command failed")
		case "InProgress", "Running":
			// Continue polling
		}
	}
}
