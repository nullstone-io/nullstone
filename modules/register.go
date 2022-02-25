package modules

import (
	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func Register(cfg api.Config, manifest *Manifest) (*types.Module, error) {
	module := &types.Module{
		OrgName:       manifest.OrgName,
		Name:          manifest.Name,
		FriendlyName:  manifest.FriendlyName,
		Description:   manifest.Description,
		IsPublic:      manifest.IsPublic,
		Category:      types.CategoryName(manifest.Category),
		Layer:         types.Layer(manifest.Layer),
		AppCategories: manifest.AppCategories,
		Type:          manifest.Type,
		ProviderTypes: manifest.ProviderTypes,
		Status:        types.ModuleStatusPublished,
	}

	client := api.Client{Config: cfg}
	if err := client.Org(module.OrgName).Modules().Create(module); err != nil {
		return nil, err
	}
	return client.Org(module.OrgName).Modules().Get(module.Name)
}
