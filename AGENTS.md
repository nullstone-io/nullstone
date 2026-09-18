# nullstone CLI — agent guide

## Releasing to production

Releases are driven by `CHANGELOG.md` and a `v<version>` git tag. There is no `make release`.

1. Every user-facing change adds a bullet under the heading at the top of `CHANGELOG.md`. A new
   release starts a new heading above the previous one: `# <major>.<minor>.<patch> (<Mon D, YYYY>)`.
   The version in the top heading IS the release version.
2. Land the code and the CHANGELOG entry on `master` through a reviewed PR. Never commit
   directly to `master`.
3. Check what is already published, then tag an up-to-date `master` with the version from the
   top CHANGELOG heading, prefixed with `v`, and push the tag:

   ```
   git ls-remote --tags origin | grep -v '\^{}' | awk -F/ '{print $NF}' | sort -V | tail -1
   git checkout master && git pull
   git tag v0.0.205          # must match the top CHANGELOG heading
   git push origin v0.0.205
   ```

4. Pushing the tag fires `.github/workflows/release.yml` (goreleaser). It builds the binaries,
   publishes the GitHub release, and pushes the Homebrew and Scoop formula updates to `master`
   as bot commits.
5. Confirm with `gh release view v0.0.205` and `gh run list --workflow=release.yml --limit 1`.

Never tag a version that has no matching CHANGELOG heading, never reuse or move a tag, and
never tag a branch other than `master`. Downstream repos (`devtools`, `terraform-provider-ns`)
consume the CLI by this tag.

Regenerate `CLI.md` with `go run ./docs` whenever commands or flags change.
