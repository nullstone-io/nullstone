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
	if u.Host == "localhost:8443" {
		u.Scheme = "http"
		u.Host = "localhost:8090"
	}

	u.Path = fmt.Sprintf("orgs/%s/stacks/%d/envs/%d/blocks/%d/activity/runs/%s",
		workspace.OrgName, workspace.StackId, workspace.EnvId, workspace.BlockId, run.Uid)
	return u.String()
}
