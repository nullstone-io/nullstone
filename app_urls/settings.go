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

// GetEnvIacOwnershipSettings links to the "IaC ownership" section of an env's settings page,
// which lists the env's events and their owning repositories.
func GetEnvIacOwnershipSettings(cfg api.Config, orgName string, stackId, envId int64) string {
	u := GetBaseUrl(cfg)
	u.Path = fmt.Sprintf("orgs/%s/stacks/%d/envs/%d/settings", orgName, stackId, envId)
	u.Fragment = "iac-ownership"
	return u.String()
}
