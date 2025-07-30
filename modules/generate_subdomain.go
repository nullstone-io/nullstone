package modules

import (
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

var (
	subdomainScaffoldTfFilename = "subdomain.tf"
	subdomainScaffoldTf         = `data "ns_subdomain" "this" {
  stack_id = data.ns_workspace.this.stack_id
  block_id = data.ns_workspace.this.block_id
}

locals {
  subdomain_dns_name = data.ns_subdomain.this.dns_name
}
`
)

func generateSubdomain(manifest *types.ModuleManifest) error {
	if manifest.Category != string(types.CategorySubdomain) {
		// We don't generate capabilities if not a subdomain module
		return nil
	}
	return generateFile(subdomainScaffoldTfFilename, subdomainScaffoldTf)
}
