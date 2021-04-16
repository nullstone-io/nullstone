package app

import "gopkg.in/nullstone-io/go-api-client.v0/types"

type Providers map[types.CategoryName]map[string]Provider

func (p Providers) Find(category types.CategoryName, type_ string) Provider {
	grp, ok := p[category]
	if !ok {
		return nil
	}
	provider, ok := grp[type_]
	if !ok {
		return nil
	}
	return provider
}
