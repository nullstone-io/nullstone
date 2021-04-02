package nullstone

type WorkspaceDeploymentInfos struct {
	client *Client
}

func (e WorkspaceDeploymentInfos) Get(stackName, blockName, envName string) (*WorkspaceDeploymentInfo, error) {
	panic("not implemented")
}
