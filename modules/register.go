package modules

import (
	"context"

	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

func Register(ctx context.Context, cfg api.Config, manifest *types.ModuleManifest) (*types.Module, error) {
	input := api.CreateModuleInput{
		Name:          manifest.Name,
		FriendlyName:  manifest.FriendlyName,
		Description:   manifest.Description,
		IsPublic:      manifest.IsPublic,
		Category:      manifest.Category,
		Subcategory:   manifest.Subcategory,
		ProviderTypes: manifest.ProviderTypes,
		Platform:      manifest.Platform,
		Subplatform:   manifest.Subplatform,
		AppCategories: manifest.AppCategories,
		Type:          manifest.Type,
		Status:        string(types.ModuleStatusPublished),
		SourceUrl:     manifest.SourceUrl,
	}

	client := api.Client{Config: cfg}
	return client.Modules().Create(ctx, manifest.OrgName, input)
}
