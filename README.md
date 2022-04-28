# Nullstone

Nullstone is a Heroku-like developer platform launched on your cloud accounts.
We offer a simple developer experience for teams that want to use Infrastructure-as-code tools like Terraform.

This repository contains code for the Nullstone CLI which is used to manage Nullstone from the command line.
This includes creating and deploying app, domains, and datastore as well as creating and managing Terraform workspaces.

## Documentation

For full documentation, visit [docs.nullstone.io](https://docs.nullstone.io).

To see how it works, visit [docs.nullstone.io/how-it-works](https://docs.nullstone.io/how-it-works).

## Community & Support

- [Public Roadmap](https://github.com/orgs/nullstone-io/projects/1/views/1) - View upcoming features.
- [GitHub Issues](https://github.com/nullstone-io/nullstone/issues) - Request new features, report bugs and errors you encounter.
- [Slack](https://join.slack.com/t/nullstone-community/signup) - Ask questions, get support, and hang out.
- Email Support - support@nullstone.io

## Quickstarts

- Static Sites/SPAs
  - React
  - Vue
- Ruby
  - Rails API
  - Rails Web App
  - Rails Sidekiq
- Python
  - Flask
  - Django
  - Celery
- Node
  - Express API
- PHP
  - Laravel
- .NET
- Go
- Elixir
  - Phoenix
- Rust

## Integrations

- [CircleCI Orb](https://github.com/nullstone-io/nullstone-orb)
- GitHub Action (In Development)
- [Go API Client](https://github.com/nullstone-io/go-api-client)

## How to install CLI

This repository contains a CLI to manage Nullstone.
This CLI works on any platform and requires no dependencies (unless you are building manually).
Nullstone currently provides easy installs for Mac and Windows (Linux coming soon).

### Homebrew (Mac)

```shell
brew tap nullstone-io/nullstone https://github.com/nullstone-io/nullstone.git
brew install nullstone
```

### Scoop (Windows)

```shell
scoop bucket add nullstone https://github.com/nullstone-io/nullstone.git
scoop install nullstone
```

### Build and install manually

This requires Go 1.17+.

```shell
go install gopkg.in/nullstone-io/nullstone.v0/nullstone
```

## Configure CLI

Visit your [Nullstone Profile](https://app.nullstone.io/profile).
Click "New API Key".
Name your API Key (usually the name of your computer or the purpose of the API Key).

Copy and run the command that is displayed in the dialog.
```shell
nullstone configure --api-key=...
```

Once you have your API Key configure, choose an org to scope CLI commands.
If you are using your personal account, use the following:
```shell
nullstone set-org <user-name>
```

If you are connecting to your organization, use the following:
```shell
nullstone set-org <org-name>
```
