package modules

import (
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"io/ioutil"
)

type generateFunc func(manifest *types.ModuleManifest) error

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
  numeric = false
  special = false
}

locals {
  tags          = data.ns_workspace.this.tags
  block_name    = data.ns_workspace.this.block_name
  resource_name = "${data.ns_workspace.this.block_ref}-${random_string.resource_suffix.result}"
}
`

	generateFns = []generateFunc{
		generateScaffold,
		generateApp,
		generateCapability,
		generateSubdomain,
		generateDomain,
	}
)

func Generate(manifest *types.ModuleManifest) error {
	for _, gfn := range generateFns {
		if err := gfn(manifest); err != nil {
			return err
		}
	}
	return nil
}

func generateScaffold(manifest *types.ModuleManifest) error {
	return generateFile(scaffoldTfFilename, baseScaffoldTf)
}

func generateFile(filename string, content string) error {
	return ioutil.WriteFile(filename, []byte(content), 0644)
}
