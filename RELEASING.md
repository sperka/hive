# Releasing Hive

Releases continue the upstream version sequence. The first fork release is
`v0.2.0`, the minor bump after `v0.1.2`. Release tags must point to reviewed commits
on `main`. Keep updater changes and installer/release changes in separate commits.

## Validate locally

```bash
mise install
mise exec -- go test -race -timeout 0 ./...
mise exec -- go vet ./...
mise exec -- shellcheck install.sh scripts/test-install.sh
bash scripts/test-install.sh
mise exec -- goreleaser check
mise exec -- goreleaser release --snapshot --clean
```

Confirm that `dist/checksums.txt` covers all four platform archives. On a matching
platform, run the built binary's `--version` and `update --help` commands.
The installer tests use local fixtures and never change the real installation.

## Publish

Update the first section of `CHANGELOG.md` for the new version and commit it.
Check that the release version matches that section before tagging.

```bash
git tag -a v0.2.0 -m "release v0.2.0"
git push --atomic origin main v0.2.0
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

Also use an older-version development build in a temporary directory to verify
an actual `hive update` against the published release. This confirms replacement,
checksum verification, and the installed version together.

## Rollback

Stop promotion if assets are missing, checksums fail, or installation, updating,
or session operations regress. Before updating the local installation, save its
binary so the first fork release has a rollback target:

```bash
cp "$(command -v hive)" ./hive.previous
```

Restore that saved binary using a temporary file in the install directory and
an atomic rename, or reinstall a known good published fork version:

```bash
curl -fsSL https://raw.githubusercontent.com/sperka/hive/main/install.sh |
  HIVE_VERSION=v0.2.0 bash
```

Use the actual known good version; the first fork release has no earlier fork
assets. Check `hive --version` and session listing after rollback. This release
does not migrate config or delete tmux sessions.

For a broken published release, remove its "latest" designation and mark it as
a prerelease while preparing a patch release. Existing installations will not
automatically downgrade: restore the saved binary or pin the installer version.
Do not rewrite published release tags.
