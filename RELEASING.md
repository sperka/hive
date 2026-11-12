# Releasing Hive

Releases continue the upstream version sequence. The current patch release is
`v0.3.2`, following the patch release `v0.3.1`. Release tags must point to
reviewed commits on `main`.

## Validate locally

```bash
mise install
mise exec -- go test -race -timeout 0 ./...
mise exec -- go vet ./...
bash -n install.sh scripts/test-install.sh
mise exec -- shellcheck install.sh scripts/test-install.sh
bash scripts/test-install.sh
mise exec -- goreleaser check
mise exec -- goreleaser release --snapshot --clean
```

Confirm that `dist/checksums.txt` covers all four platform archives and
`install.sh`. On a matching platform, run the built binary's `--version` and
`update --help` commands.
The installer tests use local fixtures and never change the real installation.
Their release version comes from the first section of `CHANGELOG.md`.

## Publish

Update the first section of `CHANGELOG.md` for the new version and commit it.
Check that the release version matches that section before tagging.

```bash
git tag -a v0.3.2 -m "release v0.3.2"
git push --atomic origin main v0.3.2
```

The tag workflow runs tests, creates a draft release, downloads the assets,
checks every checksum and archive, verifies the installer matches the tagged
source, and smoke-tests the Linux amd64 binary.
Only then does it publish the draft as the latest stable release. Failed
verification leaves the draft unpublished; inspect the failed workflow before
retrying. Do not move an already published tag.

After publication, verify the installer in a temporary directory:

```bash
release_test_dir=$(mktemp -d)
curl -fsSL https://github.com/sperka/hive/releases/latest/download/install.sh |
  HIVE_INSTALL_DIR="$release_test_dir" bash
"$release_test_dir/hive" --version
"$release_test_dir/hive" update --check
"$release_test_dir/hive" update
```

Also download the published v0.2.0, v0.3.0, and v0.3.1 archives for the current platform,
verify their checksums, and extract them into separate temporary directories.
Run each copy's `hive update --check`, `hive update`, and `hive --version` to
confirm the real upgrade to v0.3.2, checksum verification, and executable replacement. Repeat
`update --check` and `update` after the upgrade to verify follow-up updates.
Old binaries still use the API; on affected machines, bootstrap with the
v0.3.2 release-hosted installer as documented in the README.
Do not update the system installation as part of release validation.

## Rollback

Stop promotion if assets are missing, checksums fail, or installation, updating,
or session operations regress. Before updating an installation, save its binary:

```bash
cp "$(command -v hive)" ./hive.previous
```

Restore that saved binary using a temporary file in the install directory and
an atomic rename, or reinstall a known good published fork version:

```bash
curl -fsSL https://github.com/sperka/hive/releases/download/v0.3.2/install.sh |
  HIVE_VERSION=v0.3.1 bash
```

The previous release, v0.3.1, is the rollback target for v0.3.2. Check
`hive --version` and session listing after rollback. This release does not migrate
config or delete tmux sessions.

For a broken published release, remove its "latest" designation and mark it as
a prerelease while preparing a patch release. Existing installations will not
automatically downgrade: restore the saved binary or pin the installer version.
Do not rewrite published release tags.
