package runs

import (
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// GetPromotion retrieves a run configuration from the previous environment or block configuration
// moduleSourceOverride allows overriding of the returned run config based on <module-source>@<module-source-version>
func GetPromotion(cfg api.Config, workspace types.Workspace, moduleSourceOverride string) (*types.RunConfig, error) {
	client := api.Client{Config: cfg}
	newRunConfig, err := client.PromotionConfigs().Get(workspace.StackId, workspace.BlockId, workspace.EnvId, moduleSourceOverride)
	if err != nil {
		return nil, err
	}

	fillRunConfig(newRunConfig)
	return newRunConfig, nil
}
