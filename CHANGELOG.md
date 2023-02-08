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
