package modules

import (
	"gopkg.in/nullstone-io/go-api-client.v0/types"
)

var (
	appScaffoldTfFilename = "app.tf"
	appScaffoldTf         = `data "ns_app_env" "this" {
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

	appEnvVarsTfFilename = "env_vars.tf"
	appEnvVarsTf         = `variable "env_vars" {
  type        = map(string)
  default     = {}
  description = <<EOF
The environment variables to inject into the service.
These are typically used to configure a service per environment.
It is dangerous to put sensitive information in this variable because they are not protected and could be unintentionally exposed.
EOF
}

variable "secrets" {
  type        = map(string)
  default     = {}
  sensitive   = true
  description = <<EOF
The sensitive environment variables to inject into the service.
These are typically used to configure a service per environment.
EOF
}

locals {
  standard_env_vars = tomap({
    NULLSTONE_STACK         = data.ns_workspace.this.stack_name
    NULLSTONE_APP           = data.ns_workspace.this.block_name
    NULLSTONE_ENV           = data.ns_workspace.this.env_name
    NULLSTONE_VERSION       = data.ns_app_env.this.version
    NULLSTONE_COMMIT_SHA    = data.ns_app_env.this.commit_sha
    NULLSTONE_PUBLIC_HOSTS  = join(",", local.public_hosts)
    NULLSTONE_PRIVATE_HOSTS = join(",", local.private_hosts)
  })
  
  input_env_vars    = merge(local.standard_env_vars, local.cap_env_vars, var.env_vars)
  input_secrets     = merge(local.cap_secrets, var.secrets)
  input_secret_keys = nonsensitive(concat(keys(local.cap_secrets), keys(var.secrets)))
}

data "ns_env_variables" "this" {
  input_env_variables = local.input_env_vars
  input_secrets       = local.input_secrets
}

// ns_secret_keys.this is used to calculate a set of secrets to add to aws secrets manager
// The resulting "secret_keys" attribute must be known at plan time
// This doesn't need to do a full interpolation because we only care about which inputs need to be added to aws secrets manager
// ns_secret_keys.input_env_variables should contain only var.env_vars since they could contain interpolation that promotes them to sensitive
// We exclude "local.cap_env_vars" because capabilities must use "cap_secrets" to create secrets
data "ns_secret_keys" "this" {
  input_env_variables = var.env_vars
  input_secret_keys   = local.input_secret_keys
}

locals {
  secret_keys          = data.ns_secret_keys.this.secret_keys
  all_secrets          = data.ns_env_variables.this.secrets
  all_env_vars         = data.ns_env_variables.this.env_variables
  existing_secret_refs = [for key, ref in data.ns_env_variables.this.secret_refs : { name = key, valueFrom = ref }]
}
`

	appUrlsTfFilename = "urls.tf"
	appUrlsTf         = `locals {
  // Private and public URLs are shown in the Nullstone UI
  // Typically, they are created through capabilities attached to the application
  // If this module has URLs, add them here as list(string)
  additional_private_urls = []
  additional_public_urls = []

  private_urls = concat([for url in try(local.capabilities.private_urls, []) : url["url"]], local.additional_private_urls)
  public_urls  = concat([for url in try(local.capabilities.public_urls, []) : url["url"]], local.additional_public_urls)
}

locals {
  uri_matcher = "^(?:(?P<scheme>[^:/?#]+):)?(?://(?P<authority>[^/?#]*))?"
}

locals {
  authority_matcher = "^(?:(?P<user>[^@]*)@)?(?:(?P<host>[^:]*))(?:[:](?P<port>[\\d]*))?"
  // These tests are here to verify the authority_matcher regex above
  // To verify, uncomment the following lines and issue "echo 'local.tests' | terraform console"
  /*
  tests = tomap({
    "nullstone.io" : regex(local.authority_matcher, "nullstone.io"),
    "brad@nullstone.io" : regex(local.authority_matcher, "brad@nullstone.io"),
    "brad:password@nullstone.io" : regex(local.authority_matcher, "brad:password@nullstone.io"),
    "nullstone.io:9000" : regex(local.authority_matcher, "nullstone.io:9000"),
    "brad@nullstone.io:9000" : regex(local.authority_matcher, "brad@nullstone.io:9000"),
    "brad:password@nullstone.io:9000" : regex(local.authority_matcher, "brad:password@nullstone.io:9000"),
  })
  */
}

locals {
  private_hosts = [for url in local.private_urls : lookup(regex(local.authority_matcher, lookup(regex(local.uri_matcher, url), "authority")), "host")]
  public_hosts  = [for url in local.public_urls : lookup(regex(local.authority_matcher, lookup(regex(local.uri_matcher, url), "authority")), "host")]
}
`

	appOutputsTfFilename = "outputs.tf"
	appOutputsTf         = `output "private_urls" {
  value       = local.private_urls
  description = "list(string) ||| A list of URLs only accessible inside the network"
}

output "public_urls" {
  value       = local.public_urls
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

  cap_env_vars = {}
  cap_secrets  = {}
}
`

	capabilitiesTfTmplFilename = "capabilities.tf.tmpl"
	capabilitiesTfTmpl         = `{{ range . -}}
provider "ns" {
  capability_name = "{{ .Name }}"
  alias           = "{{ .TfModuleName }}"
}

module "{{ .TfModuleName }}" {
  source  = "{{ .Source }}/any"
  {{- if (ne .SourceVersion "latest") }}
  version = "{{ .SourceVersion }}"
  {{- end }}

  app_metadata = local.app_metadata
  {{ range $key, $value := .Variables -}}{{- if $value.HasValue }}
  {{ $key }} = jsondecode({{ $value.Value | to_json_string }})
  {{- end -}}{{- end }}

  providers = {
    ns = ns.{{ .TfModuleName }}
  }
}
{{ end }}
module "caps" {
  source  = "nullstone-modules/cap-merge/ns"
  modules = local.modules
}

locals {
  modules      = [
{{- range $index, $element := .ExceptNeedsDestroyed.TfModuleAddrs -}}
{{ if $index }}, {{ end }}{{ $element }}
{{- end -}}
]
  capabilities = module.caps.outputs

  cap_modules = [
{{- range $index, $element := .ExceptNeedsDestroyed }}
    {{ if $index }}, {{ end }}{
      name       = "{{ $element.Name }}"
      tfId       = "{{ $element.TfId }}"
      namespace  = "{{ $element.Namespace }}"
      env_prefix = "{{ $element.EnvPrefix }}"
      outputs    = {{ $element.TfModuleAddr }}
    }
{{- end }}
  ]
}

locals {
  cap_env_vars = merge([
    for mod in local.cap_modules : {
      for item in lookup(mod.outputs, "env", []) : "${mod.env_prefix}${item.name}" => item.value
    }
  ]...)

  cap_secrets = merge([
    for mod in local.cap_modules : {
      for item in lookup(mod.outputs, "secrets", []) : "${mod.env_prefix}${item.name}" => sensitive(item.value)
    }
  ]...)
}
`
)

func generateApp(manifest *types.ModuleManifest) error {
	if manifest.Category != string(types.CategoryApp) {
		// We don't generate capabilities if not an app module
		return nil
	}

	if err := generateFile(appScaffoldTfFilename, appScaffoldTf); err != nil {
		return err
	}
	if err := generateFile(appEnvVarsTfFilename, appEnvVarsTf); err != nil {
		return err
	}
	if err := generateFile(appUrlsTfFilename, appUrlsTf); err != nil {
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
