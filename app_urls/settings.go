package app_urls

import (
	"fmt"

	"gopkg.in/nullstone-io/go-api-client.v0"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

// blockPathSegment mirrors the UI router: apps, datastores, and domains have their own
// routes; every other block type (including subdomains) lives under /blocks.
func blockPathSegment(blockType string) string {
	switch types.BlockType(blockType) {
	case types.BlockTypeApplication:
		return "apps"
	case types.BlockTypeDatastore:
		return "datastores"
	case types.BlockTypeDomain:
		return "domains"
	default:
		return "blocks"
	}
}

// GetBlockIacOwnershipSettings links to the "IaC ownership" section of a block's settings page.
func GetBlockIacOwnershipSettings(cfg api.Config, block types.Block) string {
	u := GetBaseUrl(cfg)
	u.Path = fmt.Sprintf("orgs/%s/stacks/%d/%s/%d/settings", block.OrgName, block.StackId, blockPathSegment(block.Type), block.Id)
	u.Fragment = "iac-ownership"
	return u.String()
}

// GetEnvEventIacOwnership links to the "IaC ownership" section of an env event's edit page.
func GetEnvEventIacOwnership(cfg api.Config, event types.EnvEvent) string {
	u := GetBaseUrl(cfg)
	u.Path = fmt.Sprintf("orgs/%s/stacks/%d/envs/%d/events/%s", event.OrgName, event.StackId, event.EnvId, event.Uid)
	u.Fragment = "iac-ownership"
	return u.String()
}

// GetEnvEvents links to an env's events list, the fallback when an event could not be looked
// up by name.
func GetEnvEvents(cfg api.Config, orgName string, stackId, envId int64) string {
	u := GetBaseUrl(cfg)
	u.Path = fmt.Sprintf("orgs/%s/stacks/%d/envs/%d/events", orgName, stackId, envId)
	return u.String()
}
