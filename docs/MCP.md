# Nullstone MCP Server

The Nullstone CLI includes a built-in [MCP (Model Context Protocol)](https://modelcontextprotocol.io/) server that exposes CLI commands as tools for AI assistants like Claude.

## Getting Started

1. **Install the Nullstone CLI**

   macOS / Linux:
   ```bash
   brew tap nullstone-io/nullstone https://github.com/nullstone-io/nullstone.git
   brew install nullstone
   ```

   Windows:
   ```powershell
   scoop bucket add nullstone https://github.com/nullstone-io/nullstone.git
   scoop install nullstone
   ```

2. **Install the plugin** (Claude Code)

   ```
   /plugin marketplace add nullstone-io/nullstone
   /plugin install nullstone
   ```

3. **Log in**

   ```bash
   nullstone login
   ```

## Setup for Other Clients

### Claude Desktop

Add to your Claude Desktop configuration file:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "nullstone": {
      "command": "nullstone",
      "args": ["mcp-server"]
    }
  }
}
```

### Generic MCP Clients

The Nullstone MCP server uses **stdio transport**. Configure your client to launch `nullstone mcp-server` and communicate over stdin/stdout using the MCP JSON-RPC protocol.

## Available Tools

### Read-Only

| Tool | Description |
|------|-------------|
| `stacks_list` | List stacks in the current organization |
| `envs_list` | List environments in a stack |
| `apps_list` | List applications |
| `blocks_list` | List blocks in a stack |
| `outputs` | Get Terraform outputs for a block/environment |
| `status` | View application status (tasks, health, load balancer) |
| `profile` | View current CLI profile configuration |
| `logs` | Retrieve application logs (historical, not streaming) |

### Create / Modify

| Tool | Description |
|------|-------------|
| `stacks_new` | Create a new stack |
| `envs_new` | Create a new environment (standard or preview) |
| `envs_delete` | Delete an environment (always uses `--force`) |
| `blocks_new` | Create a new block with a module |
| `set_org` | Set the organization for the CLI profile |

### Infrastructure

| Tool | Description |
|------|-------------|
| `up` | Provision a block and its dependencies |
| `plan` | Run a Terraform plan (dry-run, auto-disapproved) |
| `apply` | Run a Terraform apply |
| `wait` | Wait for a workspace to reach a status |
| `envs_up` | Launch an entire environment |
| `envs_down` | Destroy all infrastructure in an environment |

### Deployment

| Tool | Description |
|------|-------------|
| `push` | Upload a build artifact (container image, zip, or directory) |
| `deploy` | Deploy an application version |
| `launch` | Push + deploy in a single step |

### Workspace

| Tool | Description |
|------|-------------|
| `workspaces_select` | Sync workspace Terraform state locally |

### IaC

| Tool | Description |
|------|-------------|
| `iac_test` | Test IaC files against a Nullstone stack |
| `iac_generate` | Generate IaC from a Nullstone workspace |

### Module

| Tool | Description |
|------|-------------|
| `modules_register` | Register a module in the Nullstone registry |
| `modules_publish` | Publish a new module version |
| `modules_package` | Package module contents into a tarball |

## Global Parameters

Every tool accepts these optional parameters:

| Parameter | Description |
|-----------|-------------|
| `profile` | Nullstone CLI profile name (default: `"default"`) |
| `org` | Nullstone organization name (overrides the configured org) |

## Environment Variables

The MCP server inherits the parent process environment. The following environment variables are respected:

| Variable | Description |
|----------|-------------|
| `NULLSTONE_PROFILE` | Default CLI profile |
| `NULLSTONE_ORG` | Default organization |

## Limitations

- **Interactive commands are excluded**: `exec`, `ssh`, `run`, and `modules generate` require interactive terminal input and cannot be used through MCP.
- **No streaming output**: The `logs` tool does not support `--tail` and the `status` tool does not support `--watch`, as indefinite streaming is incompatible with MCP tool results. Use `start_time`/`end_time` parameters to query log ranges.
- **`envs_delete` always forces**: Since MCP tools cannot present interactive confirmation prompts, the `envs_delete` tool always applies `--force` to skip confirmation.
- **Long-running commands**: Infrastructure commands with `--wait` (e.g., `up`, `apply`, `deploy`) may take several minutes to complete. The MCP server allows up to 30 minutes per command.
