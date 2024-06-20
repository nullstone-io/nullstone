package app_urls

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func GetRun(cfg api.Config, workspace types.Workspace, run types.Run) string {
	u := GetBaseUrl(cfg)
	u.Path = fmt.Sprintf("orgs/%s/stacks/%d/envs/%d/blocks/%d/activity/runs/%s",
		workspace.OrgName, workspace.StackId, workspace.EnvId, workspace.BlockId, run.Uid)
	return u.String()
}
