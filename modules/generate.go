package modules

import (
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"io"
	"os"
	"strings"
)

var (
	scaffoldTfFilename = "nullstone.tf"
	baseScaffoldTf     = `terraform {
  required_providers {
    ns = {
      source = "nullstone-io/ns"
    }
  }
}

data "ns_workspace" "this" {}

// Generate a random suffix to ensure uniqueness of resources
resource "random_string" "resource_suffix" {
  length  = 5
  lower   = true
  upper   = false
  number  = false
  special = false
}

locals {
  tags          = data.ns_workspace.this.tags
  block_name    = data.ns_workspace.this.block_name
  resource_name = "${data.ns_workspace.this.block_ref}-${random_string.resource_suffix.result}"
}
`
	appScaffoldTf = `
data "ns_app_env" "this" {
  stack_id = data.ns_workspace.this.stack_id
  app_id   = data.ns_workspace.this.block_id
  env_id   = data.ns_workspace.this.env_id
}

locals {
  app_version = data.ns_app_env.this.version
}
`
	subdomainScaffoldTf = `
data "ns_subdomain" "this" {
  stack_id = data.ns_workspace.this.stack_id
  block_id = data.ns_workspace.this.block_id
}

locals {
  subdomain_dns_name = data.ns_subdomain.this.dns_name
}
`
)

func Generate(manifest *Manifest) error {
	file, err := os.Create(scaffoldTfFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.WriteString(file, baseScaffoldTf); err != nil {
		return err
	}
	if strings.HasPrefix(manifest.Category, "app/") {
		if _, err := io.WriteString(file, appScaffoldTf); err != nil {
			return err
		}
	}
	if manifest.Category == types.CategorySubdomain {
		if _, err := io.WriteString(file, subdomainScaffoldTf); err != nil {
			return err
		}
	}
	return nil
}
