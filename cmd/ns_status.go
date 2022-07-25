package cmd

import (
	"fmt"
	"github.com/nullstone-io/deployment-sdk/app"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/find"
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
	awi := AppWorkspaceInfo{
		AppDetails: app.Details{
			App: application,
			Env: env,
		},
		Status:  types.WorkspaceStatusNotProvisioned,
		Version: "not-deployed",
	}

	module, err := find.Module(s.Config, awi.AppDetails.App.ModuleSource)
	if err != nil {
		return awi, err
	} else if module == nil {
		return awi, fmt.Errorf("can't find app module %s", awi.AppDetails.App.ModuleSource)
	}
	awi.AppDetails.Module = module

	client := api.Client{Config: s.Config}
	workspace, err := client.Workspaces().Get(application.StackId, application.Id, env.Id)
	if err != nil {
		return awi, err
	} else if workspace == nil {
		return awi, nil
	}
	awi.AppDetails.Workspace = workspace
	awi.Status = s.calcInfraStatus(workspace)

	appEnv, err := client.AppEnvs().Get(application.StackId, application.Id, env.Name)
	if err != nil {
		return awi, err
	}
	awi.Version = appEnv.Version
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
