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
		Subcategory:   types.SubcategoryName(manifest.Subcategory),
		ProviderTypes: manifest.ProviderTypes,
		Platform:      manifest.Platform,
		Subplatform:   manifest.Subplatform,
		AppCategories: manifest.AppCategories,
		Type:          manifest.Type,
		Status:        types.ModuleStatusPublished,
	}

	client := api.Client{Config: cfg}
	if err := client.Modules().Create(module.OrgName, module); err != nil {
		return nil, err
	}
	return client.Modules().Get(module.OrgName, module.Name)
}
