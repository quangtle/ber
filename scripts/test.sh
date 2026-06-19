#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
RUN_SERVER=false
RUN_CLIENT=false
RUN_INTEGRATION=false
RUN_E2E=false
COVERAGE=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --server) RUN_SERVER=true ;;
        --client) RUN_CLIENT=true ;;
        --integration) RUN_INTEGRATION=true ;;
        --e2e) RUN_E2E=true ;;
        --all) RUN_SERVER=true; RUN_CLIENT=true; RUN_INTEGRATION=true; RUN_E2E=true ;;
        --coverage) COVERAGE=true ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
    shift
done

if [ "$RUN_SERVER" = false ] && [ "$RUN_CLIENT" = false ] && [ "$RUN_INTEGRATION" = false ] && [ "$RUN_E2E" = false ]; then
    RUN_SERVER=true
fi

FAILED=false

if [ "$RUN_SERVER" = true ]; then
    echo -e "\n=== Server Tests ==="
    cd "$ROOT_DIR/server"
    if [ "$COVERAGE" = true ]; then
        go test -v -race -coverprofile=coverage.out -count=1 ./...
        go tool cover -html=coverage.out -o cover.html
        echo "Coverage report: cover.html"
    else
        go test -v -race -count=1 ./...
    fi
    echo "Server tests PASSED"
fi

if [ "$RUN_CLIENT" = true ]; then
    echo -e "\n=== Client Tests ==="
    CLIENT_DIR="$ROOT_DIR/client"
    BUILD_DIR="$CLIENT_DIR/build"
    mkdir -p "$BUILD_DIR"
    cd "$CLIENT_DIR"
    cmake -S . -B "$BUILD_DIR" -G Ninja -DCMAKE_BUILD_TYPE=Debug
    cmake --build "$BUILD_DIR" --target ber-client-tests --config Debug
    ctest --test-dir "$BUILD_DIR" --output-on-failure
    echo "Client tests PASSED"
fi

if [ "$RUN_INTEGRATION" = true ]; then
    echo -e "\n=== Integration Tests ==="
    cd "$ROOT_DIR/server"
    go test -v -tags=integration -count=1 ./...
    echo "Integration tests PASSED"
fi

if [ "$RUN_E2E" = true ]; then
    echo -e "\n=== E2E Tests ==="
    if [ -d "$ROOT_DIR/tests/e2e" ]; then
        cd "$ROOT_DIR/tests/e2e"
        npx playwright test
        echo "E2E tests PASSED"
    else
        echo "E2E tests directory not found, skipping"
    fi
fi

if [ "$FAILED" = true ]; then
    echo -e "\nSome tests FAILED!"
    exit 1
fi

echo -e "\nAll tests PASSED!"
