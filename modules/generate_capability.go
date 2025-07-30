package modules

import (
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

var (
	capabilityVarsTfFilename = `variables.tf`
	capabilityVarsTf         = `variable "app_metadata" {
  description = <<EOF
Nullstone automatically injects metadata from the app module into this module through this variable.
This variable is a reserved variable for capabilities.
EOF

  type    = map(string)
  default = {}
}
`
)

func generateCapability(manifest *types.ModuleManifest) error {
	if manifest.Category != string(types.CategoryCapability) {
		// We don't generate capability tf if not a capability module
		return nil
	}
	return generateFile(capabilityVarsTfFilename, capabilityVarsTf)
}
