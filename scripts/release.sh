#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
VERSION="${1:-}"
TARGET="${2:-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)}"

if [ -z "$VERSION" ]; then
    GIT_TAG="$(git describe --tags --abbrev=0 2>/dev/null || true)"
    GIT_HASH="$(git rev-parse --short HEAD 2>/dev/null || true)"
    if [[ "$GIT_TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
        VERSION="${GIT_TAG#v}"
    else
        VERSION="0.0.0-dev"
    fi
    [ -n "$GIT_HASH" ] && VERSION="${VERSION}+${GIT_HASH}"
fi

echo "=== ber Release Script ==="
echo "Version: $VERSION"
echo "Target:  $TARGET"

RELEASE_DIR="$ROOT_DIR/release"
OUT_DIR="$RELEASE_DIR/ber-$VERSION-$TARGET"
rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR"

OS="${TARGET%%-*}"
ARCH="${TARGET##*-}"
case "$OS" in
    windows) EXT=".exe" ;;
    *)       EXT="" ;;
esac

echo ""
echo "Building server for $OS/$ARCH..."
cd "$ROOT_DIR/server"
GOOS="$OS" GOARCH="$ARCH" \
    go build -ldflags="-s -w -X main.Version=$VERSION" \
    -o "$OUT_DIR/ber-server$EXT" ./cmd/ber-server/
echo "  Server: $OUT_DIR/ber-server$EXT"

if [ "$OS" = "linux" ] || [ "$OS" = "darwin" ]; then
    echo ""
    echo "Client build not yet supported for $OS, skipping."
else
    echo ""
    echo "Building client..."
    CLIENT_DIR="$ROOT_DIR/client"
    BUILD_DIR="$CLIENT_DIR/build-release"
    cd "$CLIENT_DIR"
    cmake -S . -B "$BUILD_DIR" -G Ninja -DCMAKE_BUILD_TYPE=Release
    cmake --build "$BUILD_DIR" --config Release
    cmake --install "$BUILD_DIR" --prefix "$OUT_DIR"
    echo "  Client: $OUT_DIR"
fi

cp "$ROOT_DIR/README.md" "$OUT_DIR/"

cd "$RELEASE_DIR"
if [ "$OS" = "windows" ]; then
    ARCHIVE="ber-$VERSION-$TARGET.zip"
    powershell Compress-Archive -Path "ber-$VERSION-$TARGET" -DestinationPath "$ARCHIVE" -Force
else
    ARCHIVE="ber-$VERSION-$TARGET.tar.gz"
    tar czf "$ARCHIVE" "ber-$VERSION-$TARGET"
fi

echo ""
echo "Archive: $RELEASE_DIR/$ARCHIVE"
echo ""
echo "Release complete!"
echo "Output: $OUT_DIR"
