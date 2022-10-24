# CLI Docs
Cheat sheet and reference for the Nullstone CLI.

This document contains a list of all the commands available in the Nullstone CLI along with:
- descriptions
- when to use them
- examples
- options

## apply
Runs a Terraform apply on the given block and environment. This is useful for making ad-hoc changes to your infrastructure.
Be sure to run `nullstone plan` first to see what changes will be made.

#### Usage
```shell
$ nullstone apply [--stack=<stack-name>] --block=<block-name> --env=<env-name> [options]
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--stack` | Scope this operation to a specific stack. This is only required if there are multiple blocks/apps with the same name. |  |
| `--block` | Name of the block to use for this operation | required |
| `--env` | Name of the environment to use for this operation | required |
| `--wait, -w` | Wait for the apply to complete and stream the Terraform logs to the console. |  |
| `--auto-approve` | Skip any approvals and apply the changes immediately. This requires proper permissions in the stack. |  |
| `--var` | Set variables values for the apply. This can be used to override variables defined in the module. |  |
| `--module-version` | The version of the module to apply. |  |


## apps list
Shows a list of the applications that you have access to. Set the `--detail` flag to show more details about each application.

#### Usage
```shell
$ nullstone apps list
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--detail, -d` | Use this flag to show the details for each application |  |


## blocks list
Shows a list of the blocks for the given stack. Set the `--detail` flag to show more details about each block.

#### Usage
```shell
$ nullstone blocks list --stack=<stack>
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--stack` | Set the stack to use for this operation | required |
| `--detail, -d` | Use this flag to show more details about each block |  |


## blocks new
Creates a new block with the given name and module. If the module has any connections, you can specify them using the `--connection` parameter.

#### Usage
```shell
$ nullstone blocks new --name=<name> --stack=<stack> --module=<module> [--connection=<connection>...]
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--stack` | Set the stack to use for this operation | required |
| `--name` | Provide a name for this new block | required |
| `--module` | Specify the unique name of the module to use for this block. Example: nullstone/aws-network | required |
| `--connection` | Specify any connections that this block will have to other blocks. Use the connection name as the key, and the connected block name as the value. Example: --connection network=network0 |  |


## configure
Establishes a profile and configures authentication for the CLI to use.

#### Usage
```shell
$ nullstone configure --api-key=<api-key>
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--address` | Specify the url for the Nullstone API. |  |
| `--api-key` | Configure your personal API key that will be used to authenticate with the Nullstone API. You can generate an API key from your profile page. |  |


## deploy
Deploy a new version of your code for this application. This command works in tandem with the `nullstone push` command. This command deploys the artifacts that were uploaded during the `push` command.

#### Usage
```shell
$ nullstone deploy [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--stack` | Scope this operation to a specific stack. This is only required if there are multiple blocks/apps with the same name. |  |
| `--app` | Name of the app to use for this operation |  |
| `--env` | Name of the environment to use for this operation |  |
| `--version` | Provide a label for your deployment.		If not provided, it will default to the commit sha of the repo for the current directory. |  |
| `--wait, -w` | Wait for the deploy to complete and stream the logs to the console. |  |


## envs list
Shows a list of the environments for the given stack. Set the `--detail` flag to show more details about each environment.

#### Usage
```shell
$ nullstone envs list --stack=<stack-name>
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--stack` | Set the stack to use for this operation | required |
| `--detail, -d` | Use this flag to show more details about each environment |  |


## envs new
Creates a new environment in the given stack. The environment will be created as the last environment in the pipeline. Specify the provider, region, and zone to determine where infrastructure will be provisioned for this environment.

#### Usage
```shell
$ nullstone envs new --name=<name> --stack=<stack> --provider=<provider>
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--name` | Provide a name for this new environment | required |
| `--stack` | Scope this operation to a specific stack. This is only required if there are multiple blocks/apps with the same name. |  |
| `--provider` | Provide the name of the provider to use for this environment | required |
| `--region` | Select which region to launch infrastructure for this environment. Defaults to us-east-1 for AWS and us-east1 for GCP. |  |
| `--zone` | For GCP, select the zone to launch infrastructure for this environment. Defaults to us-east1b |  |


## exec
Executes a command on a container or the virtual machine for the given application. Defaults command to '/bin/sh' which acts as opening a shell to the running container or virtual machine.

#### Usage
```shell
$ nullstone exec [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options] [command]
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--stack` | Scope this operation to a specific stack. This is only required if there are multiple blocks/apps with the same name. |  |
| `--app` | Name of the app to use for this operation |  |
| `--env` | Name of the environment to use for this operation | required |
| `--task` | Select a specific task/replica to execute the command against.		This is optional and by default will connect to a random task/replica.       	If using Kubernetes, this will select which replica of the pod to connect.       	If using ECS, this will select which task of the service to connect. |  |


## launch
This command will first upload (push) an artifact containing the source for your application. Then it will deploy it to the given environment and tail the logs for the deployment.This command is the same as running `nullstone push` followed by `nullstone deploy -w`.

#### Usage
```shell
$ nullstone launch [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--stack` | Scope this operation to a specific stack. This is only required if there are multiple blocks/apps with the same name. |  |
| `--app` | Name of the app to use for this operation |  |
| `--env` | Name of the environment to use for this operation |  |
| `--source` | The source artifact to push that contains your application's build.		For a container, specify the name of the docker image to push. This follows the same syntax as 'docker push NAME[:TAG]'.		For a serverless zip application, specify the .zip archive to push. | required |
| `--version` | Provide a label for your deployment.		If not provided, it will default to the commit sha of the repo for the current directory. |  |


## logs
Streams an application's logs to the console for the given environment. Use the start-time `-s` and end-time `-e` flags to only show logs for a given time period. Use the tail flag `-t` to stream the logs in real time.

#### Usage
```shell
$ nullstone logs [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--stack` | Scope this operation to a specific stack. This is only required if there are multiple blocks/apps with the same name. |  |
| `--app` | Name of the app to use for this operation |  |
| `--env` | Name of the environment to use for this operation |  |
| `--start-time, -s` |        Emit log events that occur after the specified start-time.        This is a golang duration relative to the time the command is issued.       Examples: '5s' (5 seconds ago), '1m' (1 minute ago), '24h' (24 hours ago)       |  |
| `--end-time, -e` |        Emit log events that occur before the specified end-time.        This is a golang duration relative to the time the command is issued.       Examples: '5s' (5 seconds ago), '1m' (1 minute ago), '24h' (24 hours ago)       |  |
| `--interval` | Set --interval to a golang duration to control how often to pull new log events.       This will do nothing unless --tail is set. The default is 1 second.       |  |
| `--tail, -t` | Set tail to watch log events and emit as they are reported.       Use --interval to control how often to query log events.       This is off by default. Unless this option is provided, this command will exit as soon as current log events are emitted. |  |


## modules generate
Generates a nullstone manifest file for your module in the current directory. You will be asked a series of questions in order to collect the information needed to describe a Nullstone module. Optionally, you can also register the module in the Nullstone registry by passing the `--register` flag.

#### Usage
```shell
$ nullstone modules generate [--register]
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--register` | Register the module in the Nullstone registry after generating the manifest file. |  |


## modules register
Registers a module in the Nullstone registry. The information in .nullstone/module.yml will be used as the details for the new module.

#### Usage
```shell
$ nullstone modules register
```

## modules publish
Publishes a new version for a module in the Nullstone registry. Provide a specific version semver using the `--version` parameter.

#### Usage
```shell
$ nullstone modules publish --version=<version>
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--version, -v` | Specify a semver version for the module.'next-patch': Uses a version that bumps the patch component of the latest module version.'next-build': Uses the latest version and appends +`build` using the short Git commit SHA. (Fails if not in a Git repository) | required |
| `--include` | Specify additional file patterns to package. By default, this command includes *.tf, *.tf.tmpl, and 'README.md'. Use this flag to package additional modules and files needed for applies. This supports file globbing detailed at https://pkg.go.dev/path/filepath#Glob |  |


## modules package
Package all the module contents for a Nullstone module into a tarball but do not publish to the registry.

#### Usage
```shell
$ nullstone modules package
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--include` | Specify additional file patterns to package. By default, this command includes *.tf, *.tf.tmpl, and 'README.md'. Use this flag to package additional modules and files needed for applies. This supports file globbing detailed at https://pkg.go.dev/path/filepath#Glob |  |


## outputs
Print all the module outputs for a given block and environment. Provide the `--sensitive` flag to include sensitive outputs in the results. For less information in an easier to read format, use the `--plain` flag.

#### Usage
```shell
$ nullstone outputs [--stack=<stack-name>] --block=<block-name> --env=<env-name> [options]
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--stack` | Scope this operation to a specific stack. This is only required if there are multiple blocks/apps with the same name. |  |
| `--block` | Name of the block to use for this operation | required |
| `--env` | Name of the environment to use for this operation | required |
| `--sensitive` | Include sensitive outputs in the results |  |
| `--plain` | Print less information about the outputs in a more readable format |  |


## plan
Run a plan for a given block and environment. This will automatically disapprove the plan and is useful for testing what a plan will do.

#### Usage
```shell
$ nullstone plan [--stack=<stack-name>] --block=<block-name> --env=<env-name> [options]
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--stack` | Scope this operation to a specific stack. This is only required if there are multiple blocks/apps with the same name. |  |
| `--block` | Name of the block to use for this operation | required |
| `--env` | Name of the environment to use for this operation | required |
| `--wait, -w` | Wait for the apply to complete and stream the Terraform logs to the console. |  |
| `--var` | Set variables values for the plan. This can be used to override variables defined in the module. |  |
| `--module-version` | Run a plan with a specific version of the module. |  |


## push
Upload (push) an artifact containing the source for your application. Specify a version semver to associate with the artifact. The version specified can be used in the deploy command to select this artifact.

#### Usage
```shell
$ nullstone push [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--stack` | Scope this operation to a specific stack. This is only required if there are multiple blocks/apps with the same name. |  |
| `--app` | Name of the app to use for this operation |  |
| `--env` | Name of the environment to use for this operation |  |
| `--source` | The source artifact to push that contains your application's build.		For a container, specify the name of the docker image to push. This follows the same syntax as 'docker push NAME[:TAG]'.		For a serverless zip application, specify the .zip archive to push. | required |
| `--version` | Provide a label for your deployment.		If not provided, it will default to the commit sha of the repo for the current directory. |  |


## set-org
Most Nullstone CLI commands require a configured nullstone organization to operate. This command will set the organization for the current profile. If you wish to set the organization per command, use the global `--org` flag instead.

#### Usage
```shell
$ nullstone set-org <org-name>
```

## ssh
SSH into a running app container or virtual machine. Use the `--forward, L` option to forward ports from remote service or hosts.

#### Usage
```shell
$ nullstone ssh [--stack=<stack-name>] --app=<app-name> --env=<env-name> [options]
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--stack` | Scope this operation to a specific stack. This is only required if there are multiple blocks/apps with the same name. |  |
| `--app` | Name of the app to use for this operation |  |
| `--env` | Name of the environment to use for this operation | required |
| `--task` | Select a specific task/replica to execute the command against.		This is optional and by default will connect to a random task/replica.       	If using Kubernetes, this will select which replica of the pod to connect.       	If using ECS, this will select which task of the service to connect. |  |
| `--forward, -L` | Use this to forward ports from host to local machine. Format: `local-port`:[`remote-host`]:`remote-port` |  |


## stacks list
Shows a list of the stacks that you have access to. Set the `--detail` flag to show more details about each stack.

#### Usage
```shell
$ nullstone stacks list
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--detail, -d` | Use this flag to show more details about each stack |  |


## stacks new
Creates a new stack with the given name and in the organization configured for the CLI.

#### Usage
```shell
$ nullstone stacks new --name=<name> --description=<description>
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--name` | The name of the stack to create. This name must be unique within the organization. | required |
| `--description` | The description of the stack to create. | required |


## status
View the status of your application and whether it is starting up, running, stopped, etc. This command shows the status of an application's tasks as well as the health of the load balancer.

#### Usage
```shell
$ nullstone status [--stack=<stack-name>] --app=<app-name> [--env=<env-name>] [options]
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--stack` | Scope this operation to a specific stack. This is only required if there are multiple blocks/apps with the same name. |  |
| `--app` | Name of the app to use for this operation |  |
| `--env` | Name of the environment to use for this operation |  |
| `--version` | Provide a label for your deployment.		If not provided, it will default to the commit sha of the repo for the current directory. |  |
| `--watch, -w` | Pass this flag in order to watch status updates in real time. Changes will be automatically displayed as they occur. |  |


## up
Launches the infrastructure for the given block and environment as well as all it's dependencies.

#### Usage
```shell
$ nullstone up [--stack=<stack-name>] --block=<block-name> --env=<env-name> [options]
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--stack` | Scope this operation to a specific stack. This is only required if there are multiple blocks/apps with the same name. |  |
| `--block` | Name of the block to use for this operation | required |
| `--env` | Name of the environment to use for this operation | required |
| `--wait, -w` | Wait for the launch to complete and stream the Terraform logs to the console. |  |
| `--var` | Set variables values for the plan. This can be used to override variables defined in the module. |  |


## version
Prints the version of the CLI.

#### Usage
```shell
nullstone -v
```

## workspaces select
Sync a given workspace's state with the current directory. Running this command will allow you to run terraform plans locally against the selected workspace.

#### Usage
```shell
$ nullstone workspaces select [--stack=<stack>] --block=<block> --env=<env>
```

#### Options
| Option | Description | |
| --- | --- | --- |
| `--stack` | Scope this operation to a specific stack. This is only required if there are multiple blocks/apps with the same name. |  |
| `--block` | Name of the block to use for this operation | required |
| `--env` | Name of the environment to use for this operation | required |


