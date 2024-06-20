package app_urls

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func GetIntentWorkflow(cfg api.Config, iw types.IntentWorkflow) string {
	u := GetBaseUrl(cfg)
	u.Path = fmt.Sprintf("orgs/%s/stacks/%d/envs/%d/activity/workflows/%d",
		iw.OrgName, iw.StackId, iw.EnvId, iw.Id)
	return u.String()
}
