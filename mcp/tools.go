package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// AllTools returns all MCP tool definitions for the Nullstone CLI.
func AllTools() []mcp.Tool {
	return []mcp.Tool{
		// Read-only
		stacksListTool(),
		envsListTool(),
		appsListTool(),
		blocksListTool(),
		outputsTool(),
		statusTool(),
		profileTool(),
		logsTool(),

		// Create/modify
		stacksNewTool(),
		envsNewTool(),
		envsDeleteTool(),
		blocksNewTool(),
		setOrgTool(),

		// Infrastructure
		upTool(),
		planTool(),
		applyTool(),
		waitTool(),
		envsUpTool(),
		envsDownTool(),

		// Deployment
		pushTool(),
		deployTool(),
		launchTool(),

		// Workspace
		workspacesSelectTool(),

		// IaC
		iacTestTool(),
		iacGenerateTool(),

		// Module
		modulesRegisterTool(),
		modulesPublishTool(),
		modulesPackageTool(),
	}
}

// withGlobalParams prepends the optional --profile and --org parameters
// that are available on every tool.
func withGlobalParams(opts ...mcp.ToolOption) []mcp.ToolOption {
	return append([]mcp.ToolOption{
		mcp.WithString("profile",
			mcp.Description("Nullstone CLI profile name. Uses 'default' profile if not specified."),
		),
		mcp.WithString("org",
			mcp.Description("Nullstone organization name. Overrides the org configured for the profile."),
		),
	}, opts...)
}

// --- Read-only tools ---

func stacksListTool() mcp.Tool {
	return mcp.NewTool("stacks_list",
		withGlobalParams(
			mcp.WithDescription("List all stacks you have access to in the current organization."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithBoolean("detail",
				mcp.Description("Show detailed information (ID, Name, Description) for each stack."),
			),
		)...,
	)
}

func envsListTool() mcp.Tool {
	return mcp.NewTool("envs_list",
		withGlobalParams(
			mcp.WithDescription("List all environments in a stack."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("stack",
				mcp.Required(),
				mcp.Description("Name of the stack."),
			),
			mcp.WithBoolean("detail",
				mcp.Description("Show detailed information for each environment."),
			),
		)...,
	)
}

func appsListTool() mcp.Tool {
	return mcp.NewTool("apps_list",
		withGlobalParams(
			mcp.WithDescription("List all applications you have access to."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithBoolean("detail",
				mcp.Description("Show detailed information for each application."),
			),
		)...,
	)
}

func blocksListTool() mcp.Tool {
	return mcp.NewTool("blocks_list",
		withGlobalParams(
			mcp.WithDescription("List all blocks in a stack."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("stack",
				mcp.Required(),
				mcp.Description("Name of the stack."),
			),
			mcp.WithBoolean("detail",
				mcp.Description("Show detailed information for each block."),
			),
		)...,
	)
}

func outputsTool() mcp.Tool {
	return mcp.NewTool("outputs",
		withGlobalParams(
			mcp.WithDescription("Retrieve Terraform outputs for a block in an environment. Returns JSON with output names, values, and types."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("stack",
				mcp.Description("Name of the stack. Only required if multiple blocks share the same name."),
			),
			mcp.WithString("block",
				mcp.Required(),
				mcp.Description("Name of the block."),
			),
			mcp.WithString("env",
				mcp.Required(),
				mcp.Description("Name of the environment."),
			),
			mcp.WithBoolean("sensitive",
				mcp.Description("Include sensitive outputs in results. Requires proper permissions."),
			),
			mcp.WithBoolean("plain",
				mcp.Description("Return simplified key-value pairs without type metadata."),
			),
		)...,
	)
}

func statusTool() mcp.Tool {
	return mcp.NewTool("status",
		withGlobalParams(
			mcp.WithDescription("View application status including task health and load balancer health. If --env is omitted, shows status across all environments."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("stack",
				mcp.Description("Name of the stack. Only required if multiple apps share the same name."),
			),
			mcp.WithString("app",
				mcp.Required(),
				mcp.Description("Name of the application."),
			),
			mcp.WithString("env",
				mcp.Description("Name of the environment. If omitted, shows status across all environments."),
			),
			mcp.WithString("version",
				mcp.Description("Filter status by deployment version."),
			),
		)...,
	)
}

func profileTool() mcp.Tool {
	return mcp.NewTool("profile",
		withGlobalParams(
			mcp.WithDescription("View the current CLI profile configuration including profile name, API address, organization, and API key status."),
			mcp.WithReadOnlyHintAnnotation(true),
		)...,
	)
}

func logsTool() mcp.Tool {
	return mcp.NewTool("logs",
		withGlobalParams(
			mcp.WithDescription("Retrieve application logs for a given environment. Returns historical logs (not streaming). Use start_time and end_time to filter by time range."),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithString("stack",
				mcp.Description("Name of the stack. Only required if multiple apps share the same name."),
			),
			mcp.WithString("app",
				mcp.Required(),
				mcp.Description("Name of the application."),
			),
			mcp.WithString("env",
				mcp.Description("Name of the environment."),
			),
			mcp.WithString("start_time",
				mcp.Description("Only show logs after this time. Go duration relative to now, e.g. '5m', '1h', '24h'."),
			),
			mcp.WithString("end_time",
				mcp.Description("Only show logs before this time. Go duration relative to now, e.g. '5m', '1h', '24h'."),
			),
		)...,
	)
}

// --- Create/modify tools ---

func stacksNewTool() mcp.Tool {
	return mcp.NewTool("stacks_new",
		withGlobalParams(
			mcp.WithDescription("Create a new stack in the current organization."),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Name of the stack. Must be unique within the organization."),
			),
			mcp.WithString("description",
				mcp.Required(),
				mcp.Description("Description of the stack."),
			),
		)...,
	)
}

func envsNewTool() mcp.Tool {
	return mcp.NewTool("envs_new",
		withGlobalParams(
			mcp.WithDescription("Create a new environment in a stack. Use --preview for preview environments. For standard environments, specify provider and region."),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Name for the new environment. For preview envs, recommend '<branch>-<pr_id>'."),
			),
			mcp.WithString("stack",
				mcp.Required(),
				mcp.Description("Name of the stack."),
			),
			mcp.WithBoolean("preview",
				mcp.Description("Create a preview environment instead of a standard pipeline environment."),
			),
			mcp.WithString("provider",
				mcp.Description("Provider name. Required for standard (non-preview) environments."),
			),
			mcp.WithString("region",
				mcp.Description("Cloud region. Defaults to us-east-1 (AWS) or us-east1 (GCP)."),
			),
			mcp.WithString("zone",
				mcp.Description("GCP zone. Defaults to us-east1b. Only used for GCP."),
			),
		)...,
	)
}

func envsDeleteTool() mcp.Tool {
	return mcp.NewTool("envs_delete",
		withGlobalParams(
			mcp.WithDescription("Delete an environment. Make sure all infrastructure has been destroyed first. Always uses --force to skip interactive confirmation."),
			mcp.WithDestructiveHintAnnotation(true),
			mcp.WithString("stack",
				mcp.Required(),
				mcp.Description("Name of the stack."),
			),
			mcp.WithString("env",
				mcp.Required(),
				mcp.Description("Name of the environment to delete."),
			),
		)...,
	)
}

func blocksNewTool() mcp.Tool {
	return mcp.NewTool("blocks_new",
		withGlobalParams(
			mcp.WithDescription("Create a new block with the given module. Optionally specify connections to other blocks."),
			mcp.WithString("stack",
				mcp.Required(),
				mcp.Description("Name of the stack."),
			),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Name for the new block."),
			),
			mcp.WithString("module",
				mcp.Required(),
				mcp.Description("Module source, e.g. 'nullstone/aws-network'."),
			),
			mcp.WithArray("connection",
				mcp.Description("Connections to other blocks. Each entry is 'connection_name=block_name', e.g. 'network=network0'."),
				mcp.WithStringItems(),
			),
		)...,
	)
}

func setOrgTool() mcp.Tool {
	return mcp.NewTool("set_org",
		withGlobalParams(
			mcp.WithDescription("Set the organization for the current CLI profile. Persists so it does not need to be specified on every command."),
			mcp.WithString("org_name",
				mcp.Required(),
				mcp.Description("The organization name to set."),
			),
		)...,
	)
}

// --- Infrastructure tools ---

func upTool() mcp.Tool {
	return mcp.NewTool("up",
		withGlobalParams(
			mcp.WithDescription("Provision infrastructure for a block and its dependencies in an environment."),
			mcp.WithString("stack",
				mcp.Description("Name of the stack. Only required if multiple blocks share the same name."),
			),
			mcp.WithString("block",
				mcp.Required(),
				mcp.Description("Name of the block."),
			),
			mcp.WithString("env",
				mcp.Required(),
				mcp.Description("Name of the environment."),
			),
			mcp.WithBoolean("wait",
				mcp.Description("Wait for provisioning to complete and stream Terraform logs."),
			),
			mcp.WithArray("var",
				mcp.Description("Override module variables. Each entry is 'key=value'."),
				mcp.WithStringItems(),
			),
		)...,
	)
}

func planTool() mcp.Tool {
	return mcp.NewTool("plan",
		withGlobalParams(
			mcp.WithDescription("Run a Terraform plan for a block in an environment. The plan is automatically disapproved (dry-run). Useful for previewing infrastructure changes."),
			mcp.WithString("stack",
				mcp.Description("Name of the stack. Only required if multiple blocks share the same name."),
			),
			mcp.WithString("block",
				mcp.Required(),
				mcp.Description("Name of the block."),
			),
			mcp.WithString("env",
				mcp.Required(),
				mcp.Description("Name of the environment."),
			),
			mcp.WithBoolean("wait",
				mcp.Description("Wait for the plan to complete and stream Terraform logs."),
			),
			mcp.WithArray("var",
				mcp.Description("Override module variables. Each entry is 'key=value'."),
				mcp.WithStringItems(),
			),
			mcp.WithString("module_version",
				mcp.Description("Run plan with a specific module version."),
			),
		)...,
	)
}

func applyTool() mcp.Tool {
	return mcp.NewTool("apply",
		withGlobalParams(
			mcp.WithDescription("Run a Terraform apply for a block in an environment. Use --auto-approve to skip approval. Run 'plan' first to preview changes."),
			mcp.WithString("stack",
				mcp.Description("Name of the stack. Only required if multiple blocks share the same name."),
			),
			mcp.WithString("block",
				mcp.Required(),
				mcp.Description("Name of the block."),
			),
			mcp.WithString("env",
				mcp.Required(),
				mcp.Description("Name of the environment."),
			),
			mcp.WithBoolean("wait",
				mcp.Description("Wait for the apply to complete and stream Terraform logs."),
			),
			mcp.WithBoolean("auto_approve",
				mcp.Description("Skip approval and apply immediately. Requires proper permissions."),
			),
			mcp.WithArray("var",
				mcp.Description("Override module variables. Each entry is 'key=value'."),
				mcp.WithStringItems(),
			),
			mcp.WithString("module_version",
				mcp.Description("Apply with a specific module version."),
			),
		)...,
	)
}

func waitTool() mcp.Tool {
	return mcp.NewTool("wait",
		withGlobalParams(
			mcp.WithDescription("Wait for a workspace to reach a specific status. Useful for waiting for infrastructure provisioning or app deployment to complete."),
			mcp.WithString("stack",
				mcp.Description("Name of the stack. Only required if multiple blocks share the same name."),
			),
			mcp.WithString("block",
				mcp.Required(),
				mcp.Description("Name of the block."),
			),
			mcp.WithString("env",
				mcp.Required(),
				mcp.Description("Name of the environment."),
			),
			mcp.WithString("for",
				mcp.Description("Status to wait for. Currently supported: 'launched'."),
			),
			mcp.WithString("timeout",
				mcp.Description("Max wait time as a Go duration. Default: '1h'. Examples: '30m', '2h'."),
			),
			mcp.WithString("approval_timeout",
				mcp.Description("Max time to wait for approval as a Go duration. Default: '15m'."),
			),
		)...,
	)
}

func envsUpTool() mcp.Tool {
	return mcp.NewTool("envs_up",
		withGlobalParams(
			mcp.WithDescription("Launch an entire environment including all apps with auto-deploy enabled. Useful for standing up preview environments."),
			mcp.WithString("stack",
				mcp.Required(),
				mcp.Description("Name of the stack."),
			),
			mcp.WithString("env",
				mcp.Required(),
				mcp.Description("Name of the environment."),
			),
		)...,
	)
}

func envsDownTool() mcp.Tool {
	return mcp.NewTool("envs_down",
		withGlobalParams(
			mcp.WithDescription("Destroy all infrastructure in an environment. Useful for tearing down preview environments."),
			mcp.WithDestructiveHintAnnotation(true),
			mcp.WithString("stack",
				mcp.Required(),
				mcp.Description("Name of the stack."),
			),
			mcp.WithString("env",
				mcp.Required(),
				mcp.Description("Name of the environment."),
			),
		)...,
	)
}

// --- Deployment tools ---

func pushTool() mcp.Tool {
	return mcp.NewTool("push",
		withGlobalParams(
			mcp.WithDescription("Upload (push) a build artifact for an application. For containers: specify docker image name. For serverless: specify .zip file. For static sites: specify directory."),
			mcp.WithString("stack",
				mcp.Description("Name of the stack. Only required if multiple apps share the same name."),
			),
			mcp.WithString("app",
				mcp.Required(),
				mcp.Description("Name of the application."),
			),
			mcp.WithString("env",
				mcp.Description("Name of the environment."),
			),
			mcp.WithString("source",
				mcp.Required(),
				mcp.Description("Source artifact to push. Docker image name, .zip file path, or directory path."),
			),
			mcp.WithString("version",
				mcp.Description("Version label for the artifact. Defaults to current git commit SHA."),
			),
		)...,
	)
}

func deployTool() mcp.Tool {
	return mcp.NewTool("deploy",
		withGlobalParams(
			mcp.WithDescription("Deploy a version of an application. Deploys artifacts previously uploaded with 'push'. Use --wait to stream deployment logs."),
			mcp.WithString("stack",
				mcp.Description("Name of the stack. Only required if multiple apps share the same name."),
			),
			mcp.WithString("app",
				mcp.Required(),
				mcp.Description("Name of the application."),
			),
			mcp.WithString("env",
				mcp.Description("Name of the environment."),
			),
			mcp.WithString("version",
				mcp.Description("Version to deploy. Defaults to current git commit SHA."),
			),
			mcp.WithBoolean("wait",
				mcp.Description("Wait for deployment to complete and stream logs."),
			),
		)...,
	)
}

func launchTool() mcp.Tool {
	return mcp.NewTool("launch",
		withGlobalParams(
			mcp.WithDescription("Push artifact and deploy in a single command. Equivalent to 'push' followed by 'deploy --wait'."),
			mcp.WithString("stack",
				mcp.Description("Name of the stack. Only required if multiple apps share the same name."),
			),
			mcp.WithString("app",
				mcp.Required(),
				mcp.Description("Name of the application."),
			),
			mcp.WithString("env",
				mcp.Description("Name of the environment."),
			),
			mcp.WithString("source",
				mcp.Required(),
				mcp.Description("Source artifact to push. Docker image name, .zip file path, or directory path."),
			),
			mcp.WithString("version",
				mcp.Description("Version label. Defaults to current git commit SHA."),
			),
		)...,
	)
}

// --- Workspace tools ---

func workspacesSelectTool() mcp.Tool {
	return mcp.NewTool("workspaces_select",
		withGlobalParams(
			mcp.WithDescription("Sync a workspace's Terraform state with the current directory. After running, you can execute terraform plan/apply locally."),
			mcp.WithString("stack",
				mcp.Description("Name of the stack. Only required if multiple blocks share the same name."),
			),
			mcp.WithString("block",
				mcp.Required(),
				mcp.Description("Name of the block."),
			),
			mcp.WithString("env",
				mcp.Required(),
				mcp.Description("Name of the environment."),
			),
		)...,
	)
}

// --- IaC tools ---

func iacTestTool() mcp.Tool {
	return mcp.NewTool("iac_test",
		withGlobalParams(
			mcp.WithDescription("Test the current directory's IaC files against a Nullstone stack and environment."),
			mcp.WithString("stack",
				mcp.Description("Name of the stack. Only required if multiple blocks share the same name."),
			),
			mcp.WithString("env",
				mcp.Required(),
				mcp.Description("Name of the environment."),
			),
		)...,
	)
}

func iacGenerateTool() mcp.Tool {
	return mcp.NewTool("iac_generate",
		withGlobalParams(
			mcp.WithDescription("Generate IaC configuration from a Nullstone workspace."),
			mcp.WithString("stack",
				mcp.Description("Name of the stack. Only required if multiple blocks share the same name."),
			),
			mcp.WithString("env",
				mcp.Required(),
				mcp.Description("Name of the environment."),
			),
			mcp.WithString("block",
				mcp.Description("Name of the block."),
			),
		)...,
	)
}

// --- Module tools ---

func modulesRegisterTool() mcp.Tool {
	return mcp.NewTool("modules_register",
		withGlobalParams(
			mcp.WithDescription("Register a module in the Nullstone registry using the .nullstone/module.yml manifest in the current directory."),
		)...,
	)
}

func modulesPublishTool() mcp.Tool {
	return mcp.NewTool("modules_publish",
		withGlobalParams(
			mcp.WithDescription("Publish a new module version to the Nullstone registry. Reads module info from .nullstone/module.yml."),
			mcp.WithString("version",
				mcp.Required(),
				mcp.Description("Semver version for the module. Special values: 'next-patch' (auto-bump patch), 'next-build' (append git SHA as build metadata)."),
			),
			mcp.WithArray("include",
				mcp.Description("Additional file patterns to package beyond *.tf, *.tf.tmpl, and README.md. Supports glob patterns."),
				mcp.WithStringItems(),
			),
		)...,
	)
}

func modulesPackageTool() mcp.Tool {
	return mcp.NewTool("modules_package",
		withGlobalParams(
			mcp.WithDescription("Package module contents into a tarball without publishing. Useful for testing before publishing."),
			mcp.WithArray("include",
				mcp.Description("Additional file patterns to package. Supports glob patterns."),
				mcp.WithStringItems(),
			),
		)...,
	)
}
