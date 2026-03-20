package appservice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/nullstone-io/deployment-sdk/app"
	"github.com/nullstone-io/deployment-sdk/logging"
	"github.com/nullstone-io/deployment-sdk/outputs"
	"gopkg.in/nullstone-io/nullstone.v0/admin"
)

const armBaseURL = "https://management.azure.com"

var (
	_ admin.Remoter = Remoter{}
)

func NewRemoter(ctx context.Context, osWriters logging.OsWriters, source outputs.RetrieverSource, appDetails app.Details) (admin.Remoter, error) {
	outs, err := outputs.Retrieve[Outputs](ctx, source, appDetails.Workspace, appDetails.WorkspaceConfig)
	if err != nil {
		return nil, err
	}
	outs.InitializeCreds(source, appDetails.Workspace)

	return Remoter{
		OsWriters: osWriters,
		Details:   appDetails,
		Infra:     outs,
	}, nil
}

type Remoter struct {
	OsWriters logging.OsWriters
	Details   app.Details
	Infra     Outputs
}

func (r Remoter) Exec(ctx context.Context, options admin.RemoteOptions, cmd []string) error {
	return r.kuduExecCommand(ctx, cmd)
}

func (r Remoter) Ssh(ctx context.Context, options admin.RemoteOptions) error {
	return KuduSsh(ctx, r.Infra, r.OsWriters)
}

func (r Remoter) Run(ctx context.Context, options admin.RunOptions, cmd []string, envVars map[string]string) error {
	return fmt.Errorf("`run` is not supported for Azure App Service")
}

// kuduExecCommand executes a command via the Kudu command API
func (r Remoter) kuduExecCommand(ctx context.Context, cmd []string) error {
	token, err := r.Infra.Remoter.GetToken(ctx, policy.TokenRequestOptions{})
	if err != nil {
		return fmt.Errorf("error getting Azure token: %w", err)
	}

	command := strings.Join(cmd, " ")
	body := map[string]interface{}{
		"command": command,
		"dir":     "/home",
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshaling exec request: %w", err)
	}

	// Use the Kudu command API via ARM proxy
	url := fmt.Sprintf("%s/resourceGroups/%s/providers/Microsoft.Web/sites/%s/extensions/api/command?api-version=2024-04-01",
		armBaseURL, r.Infra.ResourceGroup, r.Infra.SiteName)

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

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error executing command (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Output   string `json:"Output"`
		Error    string `json:"Error"`
		ExitCode int    `json:"ExitCode"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("error decoding exec result: %w", err)
	}

	stdout := r.OsWriters.Stdout()
	if result.Output != "" {
		fmt.Fprint(stdout, result.Output)
	}
	if result.Error != "" {
		fmt.Fprint(r.OsWriters.Stderr(), result.Error)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("command exited with code %d", result.ExitCode)
	}

	return nil
}
