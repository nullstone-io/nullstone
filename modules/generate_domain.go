package modules

import (
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

var (
	domainScaffoldTfFilename = "domain.tf"
	domainScaffoldTf         = `data "ns_domain" "this" {
  stack_id = data.ns_workspace.this.stack_id
  block_id = data.ns_workspace.this.block_id
}

locals {
  domain_dns_name = data.ns_domain.this.dns_name
  domain_fqdn     = "${local.domain_dns_name}."
}
`
)

func generateDomain(manifest *Manifest) error {
	if manifest.Category != string(types.CategoryDomain) {
		// We don't generate capabilities if not a domain module
		return nil
	}
	return generateFile(domainScaffoldTfFilename, domainScaffoldTf)
}
