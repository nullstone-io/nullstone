package nullstone

type WorkspaceDeploymentInfo struct {
	StackName string `json:"stackName"`
	BlockName string `json:"blockName"`
	EnvName   string `json:"envName"`
}
