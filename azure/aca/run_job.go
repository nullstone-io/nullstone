package aca

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/nullstone-io/deployment-sdk/logging"
)

// RunJob starts a Container Apps Job execution and monitors it until completion
func RunJob(ctx context.Context, osWriters logging.OsWriters, infra Outputs, cmd []string, envVars map[string]string) error {
	token, err := infra.Runner.GetToken(ctx, policy.TokenRequestOptions{})
	if err != nil {
		return fmt.Errorf("error getting Azure token: %w", err)
	}

	stdout := osWriters.Stdout()

	// Build the job start request with optional command and env var overrides
	overrides := map[string]interface{}{}
	containerOverride := map[string]interface{}{}

	if infra.MainContainerName != "" {
		containerOverride["name"] = infra.MainContainerName
	}
	if len(cmd) > 0 {
		containerOverride["command"] = cmd
	}
	if len(envVars) > 0 {
		envList := make([]map[string]string, 0, len(envVars))
		for k, v := range envVars {
			envList = append(envList, map[string]string{"name": k, "value": v})
		}
		containerOverride["env"] = envList
	}

	if len(containerOverride) > 0 {
		overrides["containerOverrides"] = []interface{}{containerOverride}
	}

	body := map[string]interface{}{}
	if len(overrides) > 0 {
		body["template"] = overrides
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshaling job start request: %w", err)
	}

	// Start the job execution
	url := fmt.Sprintf("%s/resourceGroups/%s/providers/Microsoft.App/jobs/%s/start?api-version=2024-03-01",
		armBaseURL, infra.ResourceGroup, infra.JobName)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return fmt.Errorf("error creating job start request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error starting job: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error starting job (status %d): %s", resp.StatusCode, string(respBody))
	}

	var startResult struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&startResult); err != nil {
		return fmt.Errorf("error decoding job start response: %w", err)
	}

	executionName := startResult.Name
	fmt.Fprintf(stdout, "Job execution started: %s\n", executionName)

	// Poll for job completion
	return pollJobExecution(ctx, osWriters, infra, executionName)
}

func pollJobExecution(ctx context.Context, osWriters logging.OsWriters, infra Outputs, executionName string) error {
	stdout := osWriters.Stdout()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}

		token, err := infra.Runner.GetToken(ctx, policy.TokenRequestOptions{})
		if err != nil {
			return fmt.Errorf("error getting Azure token: %w", err)
		}

		url := fmt.Sprintf("%s/resourceGroups/%s/providers/Microsoft.App/jobs/%s/executions/%s?api-version=2024-03-01",
			armBaseURL, infra.ResourceGroup, infra.JobName, executionName)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("error creating execution status request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error checking execution status: %w", err)
		}

		var execution struct {
			Properties struct {
				Status string `json:"status"`
			} `json:"properties"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&execution); err != nil {
			resp.Body.Close()
			return fmt.Errorf("error decoding execution status: %w", err)
		}
		resp.Body.Close()

		switch execution.Properties.Status {
		case "Succeeded":
			fmt.Fprintf(stdout, "Job execution completed successfully\n")
			return nil
		case "Failed":
			return fmt.Errorf("job execution failed")
		case "Running", "Processing":
			// Continue polling
		default:
			// Continue polling for unknown states
		}
	}
}
