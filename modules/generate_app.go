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
  // The runtime platform of the app; set this to one of:
  //   aws_ecs, aws_batch, aws_lambda, aws_beanstalk, aws_ec2, aws_s3, aws_eks,
  //   gcp_gke, gcp_cloudrun, gcp_cloudfunctions, gcp_composer, gcp_gce, gcp_gcs,
  //   azure_aks, azure_container_app, azure_function, azure_app_service, azure_static_web_app
  env_platform = ""

  standard_env_vars = tomap({
    NULLSTONE_STACK         = data.ns_workspace.this.stack_name
    NULLSTONE_APP           = data.ns_workspace.this.block_name
    NULLSTONE_ENV           = data.ns_workspace.this.env_name
    NULLSTONE_VERSION       = data.ns_app_env.this.version
    NULLSTONE_COMMIT_SHA    = data.ns_app_env.this.commit_sha
    NULLSTONE_PUBLIC_HOSTS  = join(",", local.public_hosts)
    NULLSTONE_PRIVATE_HOSTS = join(",", local.private_hosts)
  })

  // Variables the cloud platform provides to the app (e.g. AWS_REGION, GOOGLE_CLOUD_PROJECT)
  // Add only variables this module sets, or that the platform injects with a value known here
  cloud_env_vars = tomap({})
}

// ns_env_layout classifies secrets using keys only, so the set of secrets to create is known at plan time
// - managed_secret_keys: secrets that this module must add to the cloud secrets manager
// - unmanaged_secret_keys: references to existing secrets using {{ secret(...) }}
data "ns_env_layout" "this" {
  platform               = local.env_platform
  standard_keys          = keys(local.standard_env_vars)
  cloud_keys             = keys(local.cloud_env_vars)
  capability_env_keys    = [for e in local.capabilities.env : { capability = e.capability, name = e.name }]
  capability_secret_keys = [for s in local.capabilities.secrets : { capability = s.capability, name = s.name }]
  capability_prefixes    = local.cap_prefixes
  user_env               = var.env_vars
  user_secret_keys       = nonsensitive(keys(var.secrets))
}

// ns_env_values resolves the values to inject into the app
// - env_variables: plain environment variables
// - secrets: values for managed_secret_keys (sensitive; use ns_env_layout.managed_secret_keys in for_each)
// - unmanaged_secret_refs: ids of existing secrets referenced using {{ secret(...) }}
data "ns_env_values" "this" {
  platform            = local.env_platform
  standard            = local.standard_env_vars
  cloud               = local.cloud_env_vars
  capability_env      = local.capabilities.env
  capability_secrets  = local.capabilities.secrets
  capability_prefixes = local.cap_prefixes
  user_env            = var.env_vars
  user_secrets        = var.secrets
}

// ns_env_platform_data records the environment so Nullstone can display it
// Every key in ns_env_layout.managed_secret_keys must report where its secret lives using one of:
//   secret_ids      = { for key, secret in aws_secretsmanager_secret.app_secret : key => secret.arn }
//   k8s_secret_refs = { for key in data.ns_env_layout.this.managed_secret_keys : key => { name = "<k8s secret name>", key = key } }
data "ns_env_platform_data" "this" {
  values = data.ns_env_values.this.platform_data
}
`

	appUrlsTfFilename = "urls.tf"
	appUrlsTf         = `locals {
  // Private and public URLs are shown in the Nullstone UI
  // Typically, they are created through capabilities attached to the application
  // If this module has URLs, add them here as list(string)
  additional_private_urls = []
  additional_public_urls  = []

  private_urls = concat([for cur in local.capabilities.private_urls : cur.url], local.additional_private_urls)
  public_urls  = concat([for cur in local.capabilities.public_urls : cur.url], local.additional_public_urls)
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
  cap_modules = [
    {
      name       = ""
      tfId       = ""
      namespace  = ""
      env_prefix = ""
      outputs    = {}

      meta = {
        subcategory = ""
        platform    = ""
        subplatform = ""
        outputNames = []
      }
    }
  ]

  // cap_env_prefixes is a map indexed by tfId which points to the env_prefix in local.cap_modules
  cap_env_prefixes = tomap({
    x = ""
  })
  // cap_prefixes is a map indexed by capability name which points to the env_prefix in local.cap_modules
  cap_prefixes = tomap({
    x = ""
  })

  capabilities = {
    env = [
      {
        cap_tf_id  = "x"
        capability = "x"
        name       = "EXAMPLE_ENV"
        value      = ""
      }
    ]

    secrets = [
      {
        cap_tf_id  = "x"
        capability = "x"
        name       = "EXAMPLE_SECRET"
        value      = sensitive("")
      }
    ]

    // private_urls follows a wonky syntax so that we can send all capability outputs into the merge module
    // Terraform requires that all members be of type list(map(any))
    // They will be flattened into list(string) when we output from this module
    private_urls = [
      {
        cap_tf_id = "x"
        url       = "http://example"
      }
    ]

    // public_urls follows a wonky syntax so that we can send all capability outputs into the merge module
    // Terraform requires that all members be of type list(map(any))
    // They will be flattened into list(string) when we output from this module
    public_urls = [
      {
        cap_tf_id = "x"
        url       = "https://example.com"
      }
    ]

    // metrics allows capabilities to attach metrics to the application
    // These metrics are displayed on the Application Monitoring page
    // See https://docs.nullstone.io/extending/metrics/overview.html
    metrics = [
      {
        cap_tf_id = "x"
        name      = ""
        type      = "usage|usage-percent|duration|generic"
        unit      = ""

        mappings = jsonencode({})
      }
    ]
  }
}
`

	capabilityOutputsTfFilename = "capability_outputs.tf"
	capabilityOutputsTf         = `locals {
  // This indicates which outputs are supported by this app module
  // When adding support for a new output, add it to this list; the output will be available at "local.capabilities.<output_name>"
  capability_output_names = [
    "env",
    "secrets",
    "private_urls",
    "public_urls",
    "metrics",
  ]
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

locals {
  cap_modules = [
{{- range $index, $element := .ExceptNeedsDestroyed }}
    {{ if $index }}, {{ end }}{
      name       = "{{ $element.Name }}"
      tfId       = "{{ $element.TfId }}"
      namespace  = "{{ $element.Namespace }}"
      env_prefix = "{{ $element.EnvPrefix }}"
      outputs    = {{ $element.TfModuleAddr }}

      meta = jsondecode({{ $element.Meta | to_json_string }})
    }
{{- end }}
  ]

  cap_env_prefixes = {
    for mod in local.cap_modules : mod.tfId => mod.env_prefix
  }
  // cap_prefixes is keyed by capability name for ns_env_layout / ns_env_values
  cap_prefixes = {
    for mod in local.cap_modules : mod.name => mod.env_prefix
  }

  capabilities = {
    for outputName in local.capability_output_names : outputName => flatten([
      for mod in local.cap_modules : [ for x in lookup(mod.outputs, outputName, []) : merge({ cap_tf_id = mod.tfId, capability = mod.name }, x) ] if contains(try(mod.meta.outputNames, []), outputName)
    ])
  }
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
	if err := generateFile(capabilityOutputsTfFilename, capabilityOutputsTf); err != nil {
		return err
	}
	return generateFile(capabilitiesTfTmplFilename, capabilitiesTfTmpl)
}
