#!/bin/bash

set -euo pipefail

REPO="sperka/hive"
BINARY="hive"
INSTALL_DIR="${HIVE_INSTALL_DIR:-/usr/local/bin}"
VERSION="${HIVE_VERSION:-}"

fail() {
    echo "$*" >&2
    exit 1
}

for tool in curl tar; do
    command -v "$tool" >/dev/null 2>&1 || fail "Required command not found: $tool"
done

if command -v sha256sum >/dev/null 2>&1; then
    CHECKSUM_CMD=(sha256sum)
elif command -v shasum >/dev/null 2>&1; then
    CHECKSUM_CMD=(shasum -a 256)
else
    fail "SHA-256 verification requires sha256sum or shasum"
fi

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) fail "Unsupported architecture: $ARCH" ;;
esac

case "$OS" in
    darwin|linux) ;;
    *) fail "Unsupported OS: $OS" ;;
esac

# Get the latest stable version, unless a specific release was requested.
if [ -z "$VERSION" ]; then
    RELEASE_URL=$(curl -fsSLI -o /dev/null -w '%{url_effective}' \
        "https://github.com/$REPO/releases/latest")
    case "$RELEASE_URL" in
        "https://github.com/$REPO/releases/tag/"*)
            VERSION=${RELEASE_URL#"https://github.com/$REPO/releases/tag/"}
            ;;
        *) fail "Invalid latest release URL: $RELEASE_URL" ;;
    esac
    [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(\+[0-9A-Za-z.-]+)?$ ]] ||
        fail "Invalid stable release version: $VERSION"
fi
[[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([+-][0-9A-Za-z.-]+)?$ ]] ||
    fail "Invalid release version: $VERSION (expected v0.3.1, for example)"

echo "Installing $BINARY $VERSION for $OS/$ARCH..."

ARCHIVE="${BINARY}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION"
TMP_DIR=$(mktemp -d)
STAGED=""
USE_SUDO=0
run_install() {
    if [ "$USE_SUDO" = 1 ]; then
        sudo "$@"
    else
        "$@"
    fi
}
cleanup() {
    rm -rf "$TMP_DIR"
    if [ -n "$STAGED" ]; then
        run_install rm -f "$STAGED"
    fi
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

# Verify the archive before extracting anything.
curl -fsSL "$URL/$ARCHIVE" -o "$TMP_DIR/$ARCHIVE"
curl -fsSL "$URL/checksums.txt" -o "$TMP_DIR/checksums.txt"
EXPECTED=$(awk -v name="$ARCHIVE" '$2 == name || $2 == "*" name { print $1 }' "$TMP_DIR/checksums.txt")
[[ "$EXPECTED" =~ ^[0-9a-fA-F]{64}$ ]] || fail "Missing or invalid checksum for $ARCHIVE"
ACTUAL=$("${CHECKSUM_CMD[@]}" "$TMP_DIR/$ARCHIVE")
ACTUAL=${ACTUAL%% *}
[ "$(printf '%s' "$EXPECTED" | tr '[:upper:]' '[:lower:]')" = "$ACTUAL" ] ||
    fail "Archive checksum mismatch; Hive was not changed"
tar -xOzf "$TMP_DIR/$ARCHIVE" "$BINARY" > "$TMP_DIR/$BINARY"
[ -f "$TMP_DIR/$BINARY" ] && [ ! -L "$TMP_DIR/$BINARY" ] ||
    fail "Release archive does not contain a regular Hive binary"
[ -s "$TMP_DIR/$BINARY" ] || fail "Release archive contains an empty Hive binary"

# Create user-owned destinations without sudo where possible.
if [ ! -d "$INSTALL_DIR" ]; then
    mkdir -p "$INSTALL_DIR" 2>/dev/null || true
fi
if [ ! -w "$INSTALL_DIR" ]; then
    command -v sudo >/dev/null 2>&1 ||
        fail "Cannot write to $INSTALL_DIR; set HIVE_INSTALL_DIR to a writable directory"
    USE_SUDO=1
    run_install mkdir -p "$INSTALL_DIR"
fi

# Stage in the destination directory so the final rename is atomic.
STAGED=$(run_install mktemp "$INSTALL_DIR/.hive-install.XXXXXX")
run_install install -m 755 "$TMP_DIR/$BINARY" "$STAGED"
run_install mv -f "$STAGED" "$INSTALL_DIR/$BINARY"
STAGED=""

echo "Installed $BINARY to $INSTALL_DIR/$BINARY"
echo "Run 'hive' to start"
echo "For future updates, run 'hive update'"
case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *) echo "Add $INSTALL_DIR to your PATH to run hive" ;;
esac
