# 0.0.143 (Sep 01, 2025)
* Added support for subdomain configuration for `nullstone iac` commands.
* Improved output format for `nullstone iac test`.

# 0.0.142 (Jul 30, 2025)
* Added support for OpenTofu for managing modules.

# 0.0.141 (Jul 10, 2025)
* Record code artifact in Nullstone when performing `nullstone push|launch`.

# 0.0.140 (Mar 18, 2025)
* Added support for environment event actions (`env-launched`, `env-destroyed`) to `iac` commands.

# 0.0.139 (Mar 12, 2025)
* Fixed resolution errors on new blocks when performing `nullstone iac test`.
* Updated API client.

# 0.0.138 (Feb 19, 2025)
* Internal update to add effective and desired values to workspace config.

# 0.0.137 (Feb 17, 2025)
* Internal update to add `name` to capabilities.

# 0.0.136 (Feb 01, 2025)
* Added support for port forwarding when using `nullstone ssh` command. (`--forward <local-port>:<remote-port>`)

# 0.0.135 (Jan 06, 2025)
* Added `nullstone run` command that allows you to a start a new job/task.
* Added support for `nullstone run` to ECS/Fargate tasks and GKE jobs.

# 0.0.134 (Dec 17, 2024)
* Improve reliability of streaming deploy logs when performing a deployment.

# 0.0.133 (Dec 11, 2024)
* Added ability for users with `software_engineer` to access logs, metrics, and status.
* Added support for GCP static sites.

# 0.0.132 (Dec 08, 2024)
* Added support for pushing and deploying using Google Artifact Registry.

# 0.0.131 (Nov 21, 2024)
* Added "Deployment completed." message after streaming deployment logs.

# 0.0.130 (Nov 21, 2024)
* Created fallback when we cannot find app deployment workflow.

# 0.0.129 (Nov 12, 2024)
* Added `nullstone iac generate` command to export workspace config to an IaC file.

# 0.0.128 (Nov 05, 2024)
* Fixed nil panic when performing `nullstone iac test` without any IaC files.

# 0.0.127 (Nov 05, 2024)
* Fixed nil panic when performing `nullstone iac test` in a directory without a git repo.

# 0.0.126 (Nov 04, 2024)
* `nullstone iac test` will find IaC files with `.yml` *or* `.yaml` file extension.
* Added support for `events` in IaC files.

# 0.0.125 (Oct 25, 2024)
* `nullstone iac test` now emits previous and updated values for `module_version`, `env_variable`, `variable`, and `connection` changes.
* Eliminated diffs in `nullstone iac test` for similar values.
  * Variable values set to the same as the default
  * Connections having a missing environment id

# 0.0.124 (Oct 21, 2024)
* Fixed generation of capabilities using `workspaces select` CLI command.

# 0.0.123 (Oct 18, 2024)
* Added `nullstone iac test` command to validate Nullstone IaC files.

# 0.0.122 (Oct 01, 2024)
* Added `commitSha` when created deploy so that Nullstone can add commit info to deploy activity.
* Added detection of automation tool (e.g. CircleCI, Github Actions, Gitlab, etc.) when creating deploy.

# 0.0.121 (Jul 11, 2024)
* Updated CLI commands (`launch`, `deploy`, `plan`, `apply`, `up`, `envs up`, `envs down`) to interop with workflows.

# 0.0.120 (Apr 09, 2024)
* Added `ingress` as an option for category when generating a new module.

# 0.0.119 (Mar 25, 2024)
* Added support for AWS Batch jobs; including pushing images, deploy, and logs.

# 0.0.117 (Feb 23, 2024)
* Fixed issue where `push` and `deploy` did not find existing docker image tag due to ECR paging the image tags.

# 0.0.116 (Feb 20, 2024)
* Fixed link to run activity when running `envs up`, `plan`, `apply`, `up`.
* Added `nullstone wait` command to enable waiting for a workspace to reach `launched`.

# 0.0.115 (Feb 15, 2024)
* Provided a better mechanism for default deploy versions. This allows multiple deploys to be run from the same git commit sha.
* Always record the git commit sha on the deploy.

# 0.0.114 (Feb 15, 2024)
* Sanitizing environment name before creating in `envs new` command.

# 0.0.113 (Feb 14, 2024)
* Restoring `NULLSTONE_API_KEY` usage when running CLI commands.

# 0.0.112 (Feb 14, 2024)
* Fixed panic when user has not configured an API key.

# 0.0.111 (Jan 19, 2024)
* Fixed an issue where a Beanstalk app would sporadically use the EC2 providers for commands. (e.g. `ssh`)

# 0.0.110 (Jan 18, 2024)
* Fixed `error starting ssm session` bug when using `nullstone ssh` against AWS. (Upgraded aws sdk packages)

# 0.0.109 (Jan 17, 2024)
* Fixed bug in registration of beanstalk provider.

# 0.0.108 (Jan 16, 2024)
* Added support for `nullstone ssh` and `nullstone exec` to Elastic Beanstalk apps.

# 0.0.107 (Dec 18, 2023)
* Added log streaming support for multiple cloudwatch log groups. (Use `/*` suffix to look for multiple cloudwatch log groups)
* Moved `LogStreamer` implementations to github.com/nullstone-io/deployment-sdk.

# 0.0.106 (Nov 30, 2023)
* Added log streaming support for AWS Elastic Beanstalk.

# 0.0.105 (Nov 09, 2023)
* Added support for running `nullstone exec` for Fargate and ECS Tasks

# 0.0.104 (Jul 28, 2023)
* Fixed nil panic when specifying `--stack` that does not exist.

# 0.0.103 (Jun 07, 2023)
* Added support for `nullstone ssh` for GKE apps.
* Added support for `nullstone logs` for GKE apps.
* Fixed colorization of logs output on Windows.

# 0.0.102 (May 04, 2023)
* Fixed `nullstone --version` reporting the correct version instead of `dev`.
* Added support for `--container` when using `nullstone ssh|exec` commands for an ECS/Fargate app.

# 0.0.101 (Apr 28, 2023)
* Added `domain_fqdn` local when generating a domain module.
* Renamed `domain_name` local to `domain_dns_name` when generating a domain module.

# 0.0.100 (Apr 17, 2023)
* Added support for `cluster-namespace` on container apps.
* Dropped `service_` prefix from variables generated from `nullstone modules generate`.

# 0.0.99 (Mar 30, 2023)
* Upgraded `deployment-sdk` library to fix GCP GKE deployment support.

# 0.0.98 (Mar 29, 2023)
* Added support for `cluster-namespace` modules to module generation/creation.

# 0.0.97 (Mar 13, 2023)
* Changed `nullstone status` display format for times to emit local times using "Mon Jan _2 15:04:05 MST". 
* Changed `nullstone logs` display format for times to emit local times using "Mon Jan _2 15:04:05 MST". 

# 0.0.96 (Feb 24, 2023)
* Added `Variable.HasValue` helper function when running capability generation (`nullstone workspaces select`).

# 0.0.95 (Feb 23, 2023)
* Provide better error messages when a user is not authorized to perform an action.

# 0.0.94 (Feb 14, 2023)
* Updated the `nullstone modules generate` command to produce an updated module template.
* Support for environment variables has been updated and includes interpolation by adding the `ns_env_variables` data source.

# 0.0.93 (Feb 08, 2023)
* Nullstone modules are packaged with `.terraform.lock.hcl` and `CHANGELOG.md`. 

# 0.0.92 (Feb 02, 2023)
* Nullstone now supports the ability to queue up changes and review them before applying.
* Updated the `apply`, `plan`, and `up` commands to support pending changes.

# 0.0.91 (Jan 11, 2023)
* Added support for unversioned assets in static sites. If unversioned, assets are uploaded to the root directory.  

# 0.0.90 (Jan 06, 2023)
* Improved `nullstone ssh` to forward OS signals (e.g. `Ctrl+C`) so it does not terminate SSH tunnel.

# 0.0.89 (Dec 01, 2022)
* Fixed retrieval of module versions if none exist for a module.

# 0.0.88 (Dec 01, 2022)
* Fixed retrieval of the latest version of a module. This affected:
  * `modules publish --version=next-patch|next-build`
  * `blocks new`

# 0.0.87 (Dec 01, 2022)
* Fixed nil panic when `modules publish --version=next-patch|next-build` happens on a module with no published versions.

# 0.0.86 (Nov 07, 2022)
* Added a new `profile` command to output the current profile configuration for the CLI. Use this to help debug any unexpected results from CLI commands.
* Updated the `envs new` command to accept a new `--preview` parameter. If `--preview` is passed, the default provider and region configured for the stack will be used to configure the new preview environment.
* Added a new command `envs up` that is used to launch and deploy an entire environment.
* Added a new command `envs down` that is used to destroy an entire environment.
* Added a new command `envs delete` that is used to delete an environment once the infrastructure has been destroyed.

# 0.0.85 (Oct 21, 2022)
* Updated to account for changes in environment ordering and preview environments.
* Improved logging for runs (`up` command) and deploys (`launch` and `deploy` commands).
* Increased retry delay to 2 seconds (from 1 second) if log streaming connection fails.
* Added tracing for log streaming (use `NULLSTONE_TRACE=1`).
* Fixed nil panic when cancelling log stream that was never able to connect.

# 0.0.84 (Oct 19, 2022)
* Fixed `ssh` command for ec2 apps.
* Added helpful error message when the CLI does not support SSH for an application.

# 0.0.83 (Oct 15, 2022)
* Fixed issue when `StackName` is blank when finding an application.

# 0.0.82 (Oct 15, 2022)
* Updated `go-api-client` to adjust for changes to the Nullstone APIs.
  * Deprecated `ParentBlocks` in favor of `Connections` on Blocks/Capabilities.
  * Removed `StackName` from Blocks.

# 0.0.81 (Sep 29, 2022)
* CLI no longer prints `context canceled` when a user cancels a command.
* Updated generation of `capabilities.tf.tmpl` in application modules to generate `local.cap_modules`, `local.cap_env_vars`, & `local.cap_secrets`.
* Updated go-api-client to utilize `Namespace` and `EnvPrefix` in generation of `capabilities.tf`.

# 0.0.80 (Sep 21, 2022)
* Updated `outputs` command and deployments to use a new endpoint that references outputs from the state backend. 

# 0.0.79 (Sep 13, 2022)
* Fixed panic when running `nullstone status` on apps that do not have status support in the CLI.
* Added `--include` flag to package additional files when running `nullstone modules publish` or `nullstone modules package`.

# 0.0.78 (Aug 31, 2022)
* Fixed emitted browser URL when a plan needs approval.
* Changed `nullstone modules publish` to include `README.md` in the package.
* Changed `nullstone workspaces select` to print specific error message if unable to initialize Terraform.

# 0.0.77 (Aug 13, 2022)
* Changed deployments to run on the Nullstone servers and stream the logs through the CLI. 

# 0.0.76 (Aug 12, 2022)
* Changed `nullstone outputs` to emit like `terraform output`.
* Added `--plain` to `nullstone outputs` to emit a map of output name and output value.

# 0.0.75 (Aug 04, 2022)
* Fixed nil panic when a `plan`/`apply` received an HTTP 404 when retrieving a run configuration.

# 0.0.74 (Aug 02, 2022)
* Fixed `plan` from exiting with non-zero exit code since the plan is always "disapproved". 

# 0.0.73 (Aug 02, 2022)
* Fixed lack of error handling during `plan`/`apply`/`up` when retrieving a promotion plan.
* Switched to `gopkg.in/nullstone/go-api-client.v0` when resolving module versions by version constraint.

# 0.0.72 (Aug 02, 2022)
* When running a `plan`/`apply` with `--wait`:
  * A failed plan causes the CLI to exit with non-zero code.
  * A failed plan causes the CLI to print the error message.

# 0.0.71 (Aug 02, 2022)
* Fixed comparison of module versions when using `next-build` and `next-patch`.

# 0.0.70 (Aug 02, 2022)
* `modules publish --version=next-build|next-patch` does not consider existing versions that have build components in the version.

# 0.0.69 (Aug 01, 2022)
* Fixed `nullstone Apply` => `nullstone apply`.
* Printed Run URLs to stdout for `plan`, `apply`, and `up`.
* When running `apply`, printed message when run needs approval to proceed.

# 0.0.68 (Aug 01, 2022)
* Added `plan` command to run plans using the Nullstone engine.
* Added `apply` command to run applies (with optional `--auto-approve`) using the Nullstone engine.
* `modules publish` now emits info to stderr and emits only the new module version to stdout (if publish succeeds).
* Replaced `--version=auto` with `--version=next-patch`.
* Added `--version=next-build` that will bump the patch and append the short git commit sha as `+build`.

# 0.0.67 (Jul 29, 2022)
* Added `--wait` flag to `deploy` command that waits for the app to become healthy.
* `launch` command now performs `push`+`deploy`+`wait-healthy` instead of `push`+`deploy`+`logs`.
* Added support for elastic beanstalk apps.
* Rebuilt app providers using [nullstone-io/deployment-sdk](https://github.com/nullstone-io/deployment-sdk).

# 0.0.66 (Jul 22, 2022)
* Fixed sorting of module versions.

# 0.0.65 (Jul 21, 2022)
* Added application support for `app:container/aws/ecs:ec2`.
* Fixed `blocks new` when the selected module is an application module.
* Added `--version=auto` to `modules publish` that uses latest version and bumps the patch component.

# 0.0.64 (Jul 01, 2022)
* Fixed selection of category when running `modules generate`.

# 0.0.63 (Jun 22, 2022)
* Added support for `app:serverless/aws/lambda:container` apps.

# 0.0.62 (Jun 15, 20222)
* Updated module generation and registration to use contract-based module taxonomy.
* Renamed `modules new` to `modules register`.
* Marked `modules new` for deprecation.

# 0.0.61 (Jun 08, 2022)
* Changed `aws/lambda` provider to upload artifacts with `Content-MD5` header to work for S3 Artifacts bucket that has object lock enabled.

# 0.0.60 (Jun 07, 2022)
* Changed `aws/s3` provider to invalidate **all** content when deploying new version.

# 0.0.59 (Jun 07, 2022)
* Fixed CLI panic when `--source` directory does not exist.

# 0.0.58 (Jun 06, 2022)
* Updated terraform generation of `random_string.resource_suffix` to use `numeric` instead of deprecated `number` attribute.

# 0.0.57 (May 30, 2022)
* Added generation for `domain` modules when running `modules generate`.

# 0.0.56 (May 30, 2022)
* Added `ssh` command
  * Supports ssh for `aws-fargate` and `aws-ec2` providers.
  * Support port forwarding `--forward/-L` for `aws-ec2` provider.

# 0.0.54 (May 18, 2022)
* Fixed issue with `nullstone publish` including `v` prefix in version number.

# 0.0.53 (May 10, 2022)
* Improved `nullstone up` to set variable `Value` to `Default` if `nil`.
* Added `--var` to `nullstone up` to specify Terraform variables upon launch.

# 0.0.52 (Apr 01, 2022)
* Fixed usage of `StackId` in Deploys endpoint.

# 0.0.51 (Apr 01, 2022)
* Migrated "Update AppEnvs" endpoint to new "Create Deploy" endpoint. 

# 0.0.50 (Mar 17, 2022)
* Removed use of deprecated public module endpoints.

# 0.0.49 (Mar 01, 2022)
* Fixed `up` command:
  * If an error occurs when creating the run, do not attempt to stream the logs.
  * If we are unable to stream the logs and the user cancels (Ctrl+C/Cmd+C), then kill the process.

# 0.0.48 (Feb 28, 2022)
* Added `up` command for provisioning workspaces.
  * This command will only launch workspaces that have not provisioned yet.
  * This command comes with a `--wait` flag that will stream Terraform logs from the server.

# 0.0.47 (Feb 25, 2022)
* Enhanced `modules generate`
  * `layer` is inferred from `category` unless `category=block`.
  * Added `appCategories` when generating a capability module.
  * Generating `variables.tf` for capability modules.
  * Generating `capabilities.tf`, `capabilities.tf.tmpl`, `outputs.tf` for app modules.
* Fixed `--connection` flags in `blocks new` command.
* Updated `workspaces select` command:
  * Generating `capabilities.tf` from `capabilities.tf.tmpl`.
  * If new `ns_connection` exist locally, will prompt user to select a target for the connection.

# 0.0.46 (Feb 24, 2022)
* Updated CLI to utilize new stack-based API endpoints in Nullstone API.
* Added `connections` to `.nullstone/active-workspace.yml`.

# 0.0.45 (Feb 22, 2022)
* Fixed loading of profile so that address is set to `""` if there is no profile found.

# 0.0.44 (Not released)
* Added `aws-ec2` provider for `app/server` category with support for only `exec` command to SSH into a box.
* Added `stacks new` command.
* Added `envs new` command.
* Added `blocks list` command.
* Added `blocks new` command.
* Added `modules generate` command.
* Added `modules new` command.
* Added `modules publish` command.
* Added `workspaces select` command.

# 0.0.43 (Feb 08 2022)
* Changed all commands to use flags (e.g. `nullstone [command] --app=<app> --env=<env>`) instead of positional args (e.g. `nullstone [command] <app> <env>`).

# 0.0.42 (Jan 28 2022)
* Fixed accessing public modules from a different organization.

# 0.0.41 (Jan 27 2022)
* Added `exec` command allowing user to ssh/exec command against a container for `aws-fargate` provider.

# 0.0.40 (Jan 07 2022)
* Updated `aws-fargate` provider to use `deployer` user from fargate service instead of the cluster.

# 0.0.39 (Nov 04 2021)
* Updated `launch` to tail logs from deploy time.
* Updated `logs` to default `-s` to "now".

# 0.0.35-38 (Nov 04 2021)
* Temporary updates and fixes to allow Nullstone to deploy sensitive outputs update.

# 0.0.34 (Nov 02 2021)
* Updated API client from Nullstone changes to handle sensitive module outputs.

# 0.0.33 (Oct 19 2021)
* Upgrade to go 1.17.

# 0.0.32 (Oct 13 2021)
* Fixed resolution of app module for `status` command.

# 0.0.31 (Sep 29 2021)
* Added `site/aws-s3` provider for `app/static-site` category.

# 0.0.30 (Sep 30 2021)
* Fix panic when detecting git commit sha if there is no commit sha.

# 0.0.29 (Sep 15 2021)
* Fix panic when detecting git commit sha if there is no current git repo.

# 0.0.28 (Sep 13 2021)
* Changed `--version` (`push`, `deploy`, `launch` commands) to detect git commit sha if no version is specified.

# 0.0.27 (Aug 18 2021)
* Updated API client to handle error responses consistently.

# 0.0.26 (Jul 26 2021)
* Added `status` command for app.
* Added `status` command for app+environment.

# 0.0.25 (Jun 15 2021)
* Emitting app, stack, and environment for context when running commands.

# 0.0.24 (Jun 14 2021)
* Fixed `push`/`deploy` when no stack is specified.

# 0.0.23 (Jun 14 2021)
* Added `stacks list` command.
* Added `envs list` command.

# 0.0.22 (Jun 10 2021)
* Added `logs` command to stream logs.
* Added `cloudwatch` log provider.

# 0.0.21 (May 28 2021)
* Added `aws-ecr` provider for `app/container` category that allows `push` command.

# 0.0.20 (May 25 2021)
* Fixed retrieval of environment by name.

# 0.0.19 (May 24 2021)
* Updated retrieval of module outputs to pull the last finished run instead of the last successful run.

# 0.0.18 (May 21 2021)
* Updated API client to use ID-based endpoints.

# 0.0.17 (May 12 2021)
* Added support for `main_container_name` in `aws-fargate` outputs as a way of selecting the primary container.

# 0.0.16 (May 04 2021)
* Fixed interpretation of docker image url when there is an implicit domain.

# 0.0.15 (May 03 2021)
* Added `aws-lambda` provider for `app/serverless` category.

# 0.0.14 (Apr 29 2021)
* Fixed exit code for commands.

# 0.0.13 (Apr 29 2021)
* Use tag from source image if no image tag is specified.

# 0.0.12 (Apr 29 2021)
* Fix image tag when pushing ECR image.

# 0.0.11 (Apr 27 2021)
* Updated app version when deploying app.

# 0.0.10 (Apr 19 2021)
* Initial beta release.
* Added `aws-fargate` provider for `app/container` category.
* Added profile configuration with API key support.
