package aca

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// ExecCommand executes a command in a running Container App replica using the
// Azure Container Apps exec API
func ExecCommand(ctx context.Context, infra Outputs, container string, cmd []string) error {
	token, err := infra.Remoter.GetToken(ctx, policy.TokenRequestOptions{})
	if err != nil {
		return fmt.Errorf("error getting Azure token: %w", err)
	}

	if container == "" {
		container = infra.MainContainerName
	}
	if container == "" {
		return fmt.Errorf("no container name specified and no main container name configured")
	}

	if len(cmd) == 0 {
		cmd = []string{"/bin/sh"}
	}

	// Use the Container Apps exec API
	url := fmt.Sprintf("%s/resourceGroups/%s/providers/Microsoft.App/containerApps/%s/exec?api-version=2024-03-01",
		armBaseURL, infra.ResourceGroup, infra.ContainerAppName)

	body := map[string]interface{}{
		"containerName": container,
		"command":       cmd,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshaling exec request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return fmt.Errorf("error creating exec request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing command: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error executing command (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}
