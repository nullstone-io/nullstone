package modules

import (
	"strings"
)

var (
	appScaffoldTfFilename = "app.tf"
	appScaffoldTf         = `
data "ns_app_env" "this" {
  stack_id = data.ns_workspace.this.stack_id
  app_id   = data.ns_workspace.this.block_id
  env_id   = data.ns_workspace.this.env_id
}

locals {
  app_version = data.ns_app_env.this.version
}

locals {
  app_metadata = tomap({
    // Inject app metadata into capabilities here (e.g. security_group_name, role_name)
  })
}
`

	appOutputsTfFilename = "outputs.tf"
	appOutputsTf         = `
locals {
  // Private and public URLs are shown in the Nullstone UI
  // Typically, they are created through capabilities attached to the application
  // If this module has URLs, add them here as list(string) 
  additional_private_urls = []
  additional_public_urls  = []
}

output "private_urls" {
  value       = concat([for url in try(local.capabilities.private_urls, []) : url["url"]], local.additional_private_urls)
  description = "list(string) ||| A list of URLs only accessible inside the network"
}

output "public_urls" {
  value       = concat([for url in try(local.capabilities.public_urls, []) : url["url"]], local.additional_public_urls)
  description = "list(string) ||| A list of URLs accessible to the public"
}
`

	capabilitiesTfFilename = "capabilities.tf"
	capabilitiesTf         = `// This file is replaced by code-generation using 'capabilities.tf.tmpl'
// This file helps app module creators define a contract for what types of capability outputs are supported.
locals {
  capabilities = {
    // private_urls follows a wonky syntax so that we can send all capability outputs into the merge module
    // Terraform requires that all members be of type list(map(any))
    // They will be flattened into list(string) when we output from this module
    private_urls = [
      {
        url = ""
      }
    ]

    // public_urls follows a wonky syntax so that we can send all capability outputs into the merge module
    // Terraform requires that all members be of type list(map(any))
    // They will be flattened into list(string) when we output from this module
    public_urls = [
      {
        url = ""
      }
    ]
  }
}
`

	capabilitiesTfTmplFilename = "capabilities.tf.tmpl"
	capabilitiesTfTmpl         = `{{ range . -}}
provider "ns" {
  capability_id = {{ .Id }}
  alias         = "cap_{{ .Id }}"
}

module "{{ .TfModuleName }}" {
  source  = "{{ .Source }}/any"
  {{ if (ne .SourceVersion "latest") }}version = "{{ .SourceVersion }}"{{ end }}

  app_metadata = local.app_metadata

  {{- range $key, $value := .Variables }}
  {{ if not $value.Unused -}}
  {{ $key }} = jsondecode({{ $value.Value | to_json_string }})
  {{- end }}{{ end }}

  providers = {
    ns = ns.cap_{{ .Id}}
  }
}
{{ end }}
module "caps" {
  source  = "nullstone-modules/cap-merge/ns"
  modules = local.modules
}

locals {
  modules       = [
{{- range $index, $element := .ExceptNeedsDestroyed.TfModuleAddrs -}}
{{ if $index }}, {{ end }}{{ $element }}
{{- end -}}
]
  capabilities  = module.caps.outputs
}
`
)

func generateApp(manifest *Manifest) error {
	if !strings.HasPrefix(manifest.Category, "app/") {
		// We don't generate capabilities if not an app module
		return nil
	}

	if err := generateFile(appScaffoldTfFilename, appScaffoldTf); err != nil {
		return err
	}
	if err := generateFile(capabilitiesTfFilename, capabilitiesTf); err != nil {
		return err
	}
	if err := generateFile(appOutputsTfFilename, appOutputsTf); err != nil {
		return err
	}
	return generateFile(capabilitiesTfTmplFilename, capabilitiesTfTmpl)
}
