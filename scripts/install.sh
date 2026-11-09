#!/usr/bin/env bash
# One-command installer for qodercli-reverse on macOS / Linux.
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/sham435/qodercli-reverse/main/scripts/install.sh | bash
#   or: ./scripts/install.sh
set -euo pipefail

REPO="sham435/qodercli-reverse"
PREFIX="${PREFIX:-$HOME/.local}"
BIN_DIR="$PREFIX/bin"
BINARY="qodercli-reverse"

echo "==> qodercli-reverse installer"
echo "    prefix: $PREFIX"

# Detect platform
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac
case "$OS" in
  darwin|linux) ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

echo "    platform: $OS/$ARCH"

# Need Go to build from source (no prebuilt release assumed)
if ! command -v go >/dev/null 2>&1; then
  echo "Go is required to build from source."
  echo "Install with: brew install go   (macOS)  or  apt install golang-go  (Debian/Ubuntu)"
  exit 1
fi

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

echo "==> Cloning repository..."
git clone --depth 1 "https://github.com/$REPO.git" "$WORK/src" 2>&1 | tail -2

cd "$WORK/src/decompiled"

VERSION="v0.1.30"
COMMIT="$(git rev-parse --short HEAD)"
BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
LDFLAGS="-X github.com/alingse/qodercli-reverse/decompiled/version.Version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}"

echo "==> Building..."
GOOS="$OS" GOARCH="$ARCH" go build -ldflags "$LDFLAGS" -o "$WORK/$BINARY" .

mkdir -p "$BIN_DIR"
cp "$WORK/$BINARY" "$BIN_DIR/$BINARY"
chmod +x "$BIN_DIR/$BINARY"

echo "==> Installed: $BIN_DIR/$BINARY"
"$BIN_DIR/$BINARY" --version

# PATH hint
case ":$PATH:" in
  *":$BIN_DIR:"*) ;;
  *)
    echo ""
    echo "Add to your shell profile (~/.zshrc or ~/.bashrc):"
    echo "  export PATH=\"$BIN_DIR:\$PATH\""
    ;;
esac

echo ""
echo "Next steps:"
echo "  1. export OPENAI_API_KEY=\"sk-...\""
echo "  2. (optional) export OPENAI_BASE_URL=\"https://...\""
echo "  3. qodercli-reverse doctor"
echo "  4. qodercli-reverse --help"
