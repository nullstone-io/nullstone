package app_urls

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func GetWorkspaceWorkflow(cfg api.Config, ww types.WorkspaceWorkflow, isApp bool) string {
	u := GetBaseUrl(cfg)
	blockType := "blocks"
	if isApp {
		blockType = "apps"
	}
	u.Path = fmt.Sprintf("orgs/%s/stacks/%d/envs/%d/%s/%d/activity/workflows/%d",
		ww.OrgName, ww.StackId, ww.EnvId, blockType, ww.BlockId, ww.Id)
	return u.String()
}
