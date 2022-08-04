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
