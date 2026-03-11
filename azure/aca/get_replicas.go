package aca

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const armBaseURL = "https://management.azure.com"

type Replica struct {
	Name      string `json:"name"`
	Running   bool   `json:"running"`
	CreatedAt string `json:"createdAt"`
}

type replicaListResponse struct {
	Value []replicaResource `json:"value"`
}

type replicaResource struct {
	Name       string                   `json:"name"`
	Properties replicaResourceProperties `json:"properties"`
}

type replicaResourceProperties struct {
	CreatedTime  string              `json:"createdTime"`
	RunningState string              `json:"runningState"`
	Containers   []replicaContainer  `json:"containers"`
}

type replicaContainer struct {
	Name         string `json:"name"`
	RunningState string `json:"runningState"`
}

// GetReplicas retrieves the list of replicas for a Container App
func GetReplicas(ctx context.Context, infra Outputs) ([]Replica, error) {
	token, err := infra.Deployer.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting Azure token: %w", err)
	}

	url := fmt.Sprintf("%s/subscriptions/%s/resourceGroups/%s/providers/Microsoft.App/containerApps/%s/replicas?api-version=2024-03-01",
		armBaseURL, infra.Deployer.TenantId, infra.ResourceGroup, infra.ContainerAppName)

	// The subscription ID is not directly in Outputs — the ARM API URL is constructed using
	// the resource group which implicitly scopes to the subscription.
	// For now, we use a simplified URL pattern. The actual subscription ID would come from
	// provider config; this will be wired when the backend (Epic 2) is complete.
	url = fmt.Sprintf("%s/resourceGroups/%s/providers/Microsoft.App/containerApps/%s/replicas?api-version=2024-03-01",
		armBaseURL, infra.ResourceGroup, infra.ContainerAppName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error listing replicas: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error listing replicas (status %d): %s", resp.StatusCode, string(body))
	}

	var result replicaListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding replicas response: %w", err)
	}

	replicas := make([]Replica, 0, len(result.Value))
	for _, r := range result.Value {
		running := r.Properties.RunningState == "Running"
		replicas = append(replicas, Replica{
			Name:      r.Name,
			Running:   running,
			CreatedAt: r.Properties.CreatedTime,
		})
	}
	return replicas, nil
}
