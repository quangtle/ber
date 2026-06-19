#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BUILD_DIR="$ROOT_DIR/build"
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"

case "$OS" in
    cygwin*|mingw*|msys*)
        echo "=== ber Build Script ==="
        echo "Windows detected. Delegating to build.ps1..."
        powershell.exe -NoProfile -File "$ROOT_DIR/scripts/build.ps1" "$@"
        exit $?
        ;;
esac

mkdir -p "$BUILD_DIR"
echo "=== ber Build Script ==="
echo "OS: $OS"

if [ "${1:-}" != "--client-only" ]; then
    echo ""
    echo "Building server..."
    cd "$ROOT_DIR/server"
    go build -ldflags="-s -w" -o "$BUILD_DIR/ber-server" ./cmd/ber-server/
    echo "  $BUILD_DIR/ber-server"
    go build -ldflags="-s -w" -o "$BUILD_DIR/ber-desktop" ./cmd/ber-desktop/
    echo "  $BUILD_DIR/ber-desktop"
fi

if [ "${1:-}" != "--server-only" ]; then
    echo ""
    echo "Building client..."
    mkdir -p "$BUILD_DIR/client"
    cd "$ROOT_DIR/client"

    if [ -d "/usr/local/opt/qt@6" ]; then
        QT_DIR="/usr/local/opt/qt@6"
    elif [ -d "/usr/lib/x86_64-linux-gnu/cmake/Qt6" ]; then
        QT_DIR="/usr/lib/x86_64-linux-gnu"
    else
        QT_DIR="/opt/Qt/6.8.2/gcc_64"
    fi

    cmake -S . -B "$BUILD_DIR/client" -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_PREFIX_PATH="$QT_DIR"
    cmake --build "$BUILD_DIR/client" --config Release
    echo "  $BUILD_DIR/client/ber-client"
fi

echo ""
echo "Build complete!"
