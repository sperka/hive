# Releasing Hive

Releases continue the upstream version sequence: `v0.1.2`, the first fork release
`v0.2.0`, and the dashboard release `v0.3.0`. Release tags must point to reviewed
commits on `main`. Consolidate the dashboard and v0.3.0 release preparation into
one commit above `v0.2.0`.

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

Confirm that `dist/checksums.txt` covers all four platform archives. On a matching
platform, run the built binary's `--version` and `update --help` commands.
The installer tests use local fixtures and never change the real installation.
Their release version comes from the first section of `CHANGELOG.md`.

## Publish

Update the first section of `CHANGELOG.md` for the new version and commit it.
Check that the release version matches that section before tagging.

```bash
git tag -a v0.3.0 -m "release v0.3.0"
git push --atomic origin main v0.3.0
```

The tag workflow runs tests, creates a draft release, downloads the assets,
checks every checksum and archive, and smoke-tests the Linux amd64 binary.
Only then does it publish the draft as the latest stable release. Failed
verification leaves the draft unpublished; inspect the failed workflow before
retrying. Do not move an already published tag.

After publication, verify the installer in a temporary directory:

```bash
release_test_dir=$(mktemp -d)
curl -fsSL https://raw.githubusercontent.com/sperka/hive/main/install.sh |
  HIVE_INSTALL_DIR="$release_test_dir" bash
"$release_test_dir/hive" --version
"$release_test_dir/hive" update --check
"$release_test_dir/hive" update
```

Also download the published v0.2.0 archive for the current platform, verify its
checksum, and extract it into a separate temporary directory. Run that copy's
`hive update --check`, `hive update`, and `hive --version` to confirm the real
v0.2.0-to-v0.3.0 upgrade, checksum verification, and executable replacement.
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
curl -fsSL https://raw.githubusercontent.com/sperka/hive/main/install.sh |
  HIVE_VERSION=v0.2.0 bash
```

The previous release, v0.2.0, is the rollback target for v0.3.0. Check
`hive --version` and session listing after rollback. This release does not migrate
config or delete tmux sessions.

For a broken published release, remove its "latest" designation and mark it as
a prerelease while preparing a patch release. Existing installations will not
automatically downgrade: restore the saved binary or pin the installer version.
Do not rewrite published release tags.
