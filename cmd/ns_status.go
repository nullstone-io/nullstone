package cmd

import (
	"context"
	"fmt"

	"github.com/nullstone-io/deployment-sdk/app"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

type AppWorkspaceInfo struct {
	AppDetails app.Details
	Status     string
	Version    string
}

type NsStatus struct {
	Config api.Config
}

func (s NsStatus) GetAppWorkspaceInfo(application *types.Application, env *types.Environment) (AppWorkspaceInfo, error) {
	ctx := context.TODO()
	apiClient := api.Client{Config: s.Config}
	infraDetails, err := apiClient.WorkspaceInfraDetails().Get(ctx, application.StackId, application.Id, env.Id, false)
	if err != nil {
		return AppWorkspaceInfo{}, err
	} else if infraDetails == nil {
		return AppWorkspaceInfo{}, fmt.Errorf("Application Workspace (%s/%s) does not exist.", application.Name, env.Name)
	}

	awi := AppWorkspaceInfo{
		AppDetails: app.Details{
			App:             application,
			Env:             env,
			Workspace:       &infraDetails.Workspace,
			WorkspaceConfig: &infraDetails.WorkspaceConfig,
			Module:          &infraDetails.Module,
		},
		Status:  types.WorkspaceStatusNotProvisioned,
		Version: "not-deployed",
	}

	awi.Status = s.calcInfraStatus(awi.AppDetails.Workspace)

	appEnv, err := apiClient.AppEnvs().Get(ctx, application.StackId, application.Id, env.Name)
	if err != nil {
		return awi, err
	} else if appEnv != nil {
		awi.Version = appEnv.Version
	}
	if awi.Version == "" || awi.Status == types.WorkspaceStatusNotProvisioned || awi.Status == "creating" {
		awi.Version = "not-deployed"
	}

	return awi, nil
}

func (s NsStatus) calcInfraStatus(workspace *types.Workspace) string {
	if workspace == nil {
		return types.WorkspaceStatusNotProvisioned
	}
	if workspace.ActiveRun == nil {
		return workspace.Status
	}
	switch workspace.ActiveRun.Status {
	default:
		return workspace.Status
	case types.RunStatusNeedsApproval:
		return "needs-approval"
	case types.RunStatusResolving:
	case types.RunStatusInitializing:
	case types.RunStatusAwaiting:
	case types.RunStatusRunning:
	}
	if workspace.ActiveRun.IsDestroy {
		return "destroying"
	}
	if workspace.Status == types.WorkspaceStatusNotProvisioned {
		return "creating"
	}
	return "updating"
}
