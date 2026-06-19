# Release Process

## Overview

The release process produces platform-specific executables for `ber-server` and `ber-client`, along with the web UI embedded in the server binary.  
Run a single script to produce a `release/` directory with ready-to-distribute archives.

---

## Prerequisites

- All [development dependencies](development.md) installed
- A valid C++ toolchain (MSVC 2022 on Windows, GCC on Linux, Clang on macOS)
- `git` available on `PATH`
- (Optional) `zip` and `tar` for archive creation

---

## One-Command Release

```powershell
# Windows
.\scripts\release.ps1

# Unix
./scripts/release.sh
```

This will:
1. Detect the current platform and architecture
2. Build the server binary (statically linked where possible)
3. Build the client executable (with runtime DLLs bundled on Windows)
4. Embed the web UI into the server binary
5. Run all tests (fails on test failure)
6. Package everything into a compressed archive
7. Place output in `release/ber-<version>-<os>-<arch>/`

---

## Versioning

ber follows [Semantic Versioning 2.0](https://semver.org/). The version is derived from the most recent `git` tag matching `v*`:

```
v1.2.3 → version = "1.2.3"
```

If no tag is found, the script falls back to `0.0.0-dev` and appends the commit short hash.

### Creating a Release Tag

```powershell
git tag -a v1.0.0 -m "v1.0.0 — Initial release"
git push origin v1.0.0
```

The CI pipeline (see `ci/.github/workflows/release.yml`) automatically builds all platforms when a tag is pushed.

---

## Output Structure

```
release/
└── ber-1.0.0-windows-amd64/
    ├── ber-server.exe
    ├── ber-client.exe
    ├── Qt6Core.dll
    ├── Qt6Gui.dll
    ├── Qt6Widgets.dll
    ├── Qt6Multimedia.dll
    ├── Qt6Network.dll
    ├── ... (other Qt DLLs)
    ├── platforms/
    │   └── qwindows.dll
    ├── styles/
    │   └── ... (Qt style plugins)
    ├── LICENSE
    └── README.txt
```

On Linux/macOS the archive uses a `.tar.gz` bundle instead of `.zip`.

---

## Platform Build Matrix

| Target | OS | Arch | Server Build | Client Build |
|--------|----|------|-------------|--------------|
| windows-amd64 | Windows 10+ | x86_64 | `GOOS=windows GOARCH=amd64 go build` | MSVC 2022, Qt6 x64 |
| linux-amd64 | Linux (glibc 2.28+) | x86_64 | `GOOS=linux GOARCH=amd64 go build` | GCC 13+, Qt6 x64 |
| linux-arm64 | Linux (Raspberry Pi OS, glibc) | ARM64 | `GOOS=linux GOARCH=arm64 go build` | Web UI only (no native client build) |
| darwin-amd64 | macOS 12+ | x86_64 | `GOOS=darwin GOARCH=amd64 go build` | Clang 16+, Qt6 x64 |
| darwin-arm64 | macOS 14+ (Apple Silicon) | ARM64 | `GOOS=darwin GOARCH=arm64 go build` | Clang 16+, Qt6 arm64 |

---

## Manual Build Steps

If you want to build individual components without the release script:

### Server

```powershell
cd server
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w -X main.Version=1.0.0" -o ..\release\ber-server.exe .\cmd\ber-server\
```

### Client

```powershell
cd client
cmake -S . -B build -G Ninja `
    -DCMAKE_BUILD_TYPE=Release `
    -DCMAKE_PREFIX_PATH="C:\Qt\6.6.0\msvc2022_64"
cmake --build build --config Release
# Copy binary + dependencies
cmake --install build --prefix ..\release\ber-1.0.0-windows-amd64
```

### Web UI (embedded — no separate step)

The `server/web/` directory is embedded at compile time. No bundler or build step is needed.

---

## Verification Checklist

Before publishing a release:

- [ ] All tests pass: `.\scripts\test.ps1 -All`
- [ ] Server starts and serves web UI: `ber-server.exe --library .\testdata` → browse to http://localhost:8080
- [ ] Client connects to server and plays a video
- [ ] Static analysis clean: `golangci-lint run ./...`
- [ ] Binaries are stripped of debug symbols (`-ldflags="-s -w"` for Go; `strip` for C++)
- [ ] Version string is correct: `ber-server.exe --version`
- [ ] Fresh environment test: run the release archive on a clean machine

---

## CI / CD Publishing

When a tag `v*` is pushed, the auto-build workflow in `ci/.github/workflows/release.yml`:

1. Builds server + client for all matrix targets
2. Runs full test suite
3. Creates platform-specific archives
4. Creates a GitHub Release with the archives attached
5. (Future) Pushes to a package registry or download site

---

## Script Reference

| Script | Purpose |
|--------|---------|
| `scripts/build.ps1` | Build server + client for current platform |
| `scripts/release.ps1` | Build, bundle, archive for distribution |
| `scripts/test.ps1` | Run all tests with optional flags |
| `scripts/generate-fixtures.ps1` | Generate small test media files |
