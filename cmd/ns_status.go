package cmd

import (
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/app"
)

type AppWorkspaceInfo struct {
	AppDetails app.Details
	Status    string
	Version   string
}

type NsStatus struct {
	Config api.Config
}

func (s NsStatus) GetAppWorkspaceInfo(application *types.Application, env *types.Environment) (AppWorkspaceInfo, error) {
	awi := AppWorkspaceInfo{
		AppDetails: app.Details {
			App: application,
			Env: env,
		},
		Status: types.WorkspaceStatusNotProvisioned,
		Version: "not-deployed",
	}

	client := api.Client{Config: s.Config}
	workspace, err := client.Workspaces().Get(application.StackId, application.Id, env.Id)
	if err != nil {
		return awi, err
	} else if workspace == nil {
		return awi, nil
	}
	awi.AppDetails.Workspace = workspace
	awi.Status = s.calcInfraStatus(workspace)

	appEnv, err := client.AppEnvs().Get(application.Id, env.Name)
	if err != nil {
		return awi, err
	}
	version := appEnv.Version
	if version == "" || awi.Status == types.WorkspaceStatusNotProvisioned || awi.Status == "creating" {
		version = "not-deployed"
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