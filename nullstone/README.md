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
