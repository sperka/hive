#!/bin/bash
set -euo pipefail

PROJECT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
TEST_DIR=$(mktemp -d)
trap 'rm -rf "$TEST_DIR"' EXIT
mkdir -p "$TEST_DIR/bin" "$TEST_DIR/fixtures" "$TEST_DIR/destination" "$TEST_DIR/tmp"
export HIVE_INSTALL_TEST_DIR="$TEST_DIR"

# Only curl and uname are substituted. Downloads must use the fork's URLs.
cat > "$TEST_DIR/bin/curl" <<'SH'
#!/bin/bash
set -euo pipefail
url=""
output=""
while [ "$#" -gt 0 ]; do
    case "$1" in
        -o) output=$2; shift 2 ;;
        -*) shift ;;
        *) url=$1; shift ;;
    esac
done
printf '%s\n' "$url" >> "$HIVE_INSTALL_TEST_DIR/requests"
case "$url" in
    https://api.github.com/repos/sperka/hive/releases/latest)
        if [ "${HIVE_TEST_FAIL:-}" = api ]; then exit 22; fi
        printf '{"tag_name": "v0.2.0"}\n'
        ;;
    https://github.com/sperka/hive/releases/download/v0.2.0/*)
        name=${url##*/}
        if [ "${HIVE_TEST_FAIL:-}" = archive ] && [ "$name" != checksums.txt ]; then exit 22; fi
        cp "$HIVE_INSTALL_TEST_DIR/fixtures/$name" "$output"
        ;;
    *) echo "Unexpected URL: $url" >&2; exit 1 ;;
esac
SH
cat > "$TEST_DIR/bin/uname" <<'SH'
#!/bin/bash
case "$1" in
    -s) echo "${HIVE_TEST_OS:-Linux}" ;;
    -m) echo "${HIVE_TEST_ARCH:-aarch64}" ;;
esac
SH
chmod +x "$TEST_DIR/bin/"*

printf '#!/bin/sh\necho "hive 0.2.0"\n' > "$TEST_DIR/fixtures/hive"
chmod +x "$TEST_DIR/fixtures/hive"
for os in linux darwin; do
    for arch in amd64 arm64; do
        tar -czf "$TEST_DIR/fixtures/hive_${os}_${arch}.tar.gz" -C "$TEST_DIR/fixtures" hive
    done
done
if command -v sha256sum >/dev/null 2>&1; then
    CHECKSUM_CMD=(sha256sum)
else
    CHECKSUM_CMD=(shasum -a 256)
fi
(
    cd "$TEST_DIR/fixtures"
    "${CHECKSUM_CMD[@]}" ./*.tar.gz | sed 's| \./| |' > checksums.txt
)
cp "$TEST_DIR/fixtures/checksums.txt" "$TEST_DIR/checksums.good"

run_installer() {
    PATH="$TEST_DIR/bin:$PATH" TMPDIR="$TEST_DIR/tmp" \
        HIVE_INSTALL_DIR="$TEST_DIR/destination" bash "$PROJECT_DIR/install.sh"
}

assert_clean() {
    [ -z "$(find "$TEST_DIR/tmp" -mindepth 1 -print -quit)" ]
    [ -z "$(find "$TEST_DIR/destination" -name '.hive-install.*' -print -quit)" ]
}

for os in Linux Darwin; do
    for arch in x86_64 aarch64 arm64; do
        HIVE_TEST_OS="$os" HIVE_TEST_ARCH="$arch" run_installer
        [ "$("$TEST_DIR/destination/hive" --version)" = "hive 0.2.0" ]
        assert_clean
    done
done

# A pinned install skips the latest-release API request.
: > "$TEST_DIR/requests"
HIVE_VERSION=v0.2.0 run_installer
if grep -q 'api.github.com' "$TEST_DIR/requests"; then
    echo "Pinned install queried latest release" >&2
    exit 1
fi

# A new install creates a user-writable destination when it does not exist.
rm -r "$TEST_DIR/destination"
run_installer
[ -x "$TEST_DIR/destination/hive" ]
assert_clean

assert_failure_preserves_binary() {
    printf 'original binary\n' > "$TEST_DIR/destination/hive"
    if run_installer; then
        echo "Expected installer failure" >&2
        exit 1
    fi
    [ "$(cat "$TEST_DIR/destination/hive")" = "original binary" ]
    assert_clean
}

HIVE_TEST_FAIL=api assert_failure_preserves_binary
HIVE_TEST_FAIL=archive assert_failure_preserves_binary
HIVE_VERSION=../../invalid assert_failure_preserves_binary
HIVE_TEST_ARCH=riscv64 assert_failure_preserves_binary
HIVE_TEST_OS=Windows assert_failure_preserves_binary
sed 's/^[0-9a-f]*/0000000000000000000000000000000000000000000000000000000000000000/' \
    "$TEST_DIR/checksums.good" > "$TEST_DIR/fixtures/checksums.txt"
assert_failure_preserves_binary
: > "$TEST_DIR/fixtures/checksums.txt"
assert_failure_preserves_binary

echo "Installer tests passed"
