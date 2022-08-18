package runs

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"net/url"
	"strings"
)

func GetBrowserUrl(cfg api.Config, workspace types.Workspace, run types.Run) string {
	u, err := url.Parse(cfg.BaseAddress)
	if err != nil {
		u = &url.URL{Scheme: "https", Host: "app.nullstone.io"}
	}
	u.Host = strings.Replace(u.Host, "api", "app", 1)

	u.Path = fmt.Sprintf("orgs/%s/stacks/%d/blocks/%d/activity", workspace.OrgName, workspace.StackId, workspace.BlockId)
	q := url.Values{}
	q.Set("runUid", run.Uid.String())
	q.Set("env", fmt.Sprintf("%d", workspace.EnvId))
	u.RawQuery = q.Encode()
	return u.String()
}
