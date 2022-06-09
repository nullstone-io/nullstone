package app

import "gopkg.in/nullstone-io/go-api-client.v0/types"

type Providers map[types.ModuleContractName]Provider

func (p Providers) Find(curModule types.Module) Provider {
	if len(curModule.ProviderTypes) <= 0 {
		return nil
	}

	// NOTE: We are matching app modules, so category is redundant
	//   However, this should guard against non-app modules trying to use these app providers
	curContract := types.ModuleContractName{
		Category:    string(curModule.Category),
		Subcategory: string(curModule.Subcategory),
		// TODO: Enforce module provider can only contain one and only one provider type
		Provider:    curModule.ProviderTypes[0],
		Platform:    curModule.Platform,
		Subplatform: curModule.Subplatform,
	}
	for k, v := range p {
		if k.Match(curContract) {
			return v
		}
	}

	return nil
}
