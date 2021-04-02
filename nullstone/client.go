package nullstone

type Client struct {
	Address string
	ApiKey  string
	OrgName string
}

func (c *Client) Apps() Apps {
	return Apps{client: c}
}

func (c *Client) WorkspaceDeploymentInfos() WorkspaceDeploymentInfos {
	return WorkspaceDeploymentInfos{client: c}
}
