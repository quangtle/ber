# Development Environment Setup

## Prerequisites

### For the Server (ber-server)

| Tool | Version | Download |
|------|---------|----------|
| Go | 1.22+ | `winget install GoLang.Go` |
| FFmpeg | 6.0+ | `winget install Gyan.FFmpeg` |

### For the Client (ber-client)

| Tool | Version | Download |
|------|---------|----------|
| CMake | 3.28+ | `winget install Kitware.CMake` |
| C++ compiler | MSVC 2022 (Windows), GCC 13+ (Linux), Clang 16+ (macOS) | See VS setup below |
| Qt6 | 6.6+ (with Multimedia module) | See Qt setup below |
| Ninja | 1.11+ | `winget install Ninja-build.Ninja` |

### For Testing & Quality

| Tool | Version | Purpose |
|------|---------|---------|
| golangci-lint | 1.56+ | Go linting |
| gofumpt | latest | Go formatting |
| clang-format | 17+ | C++ formatting |
| clang-tidy | 17+ | C++ static analysis |
| Catch2 | 3.5+ | C++ unit testing (auto-downloaded via CMake FetchContent) |

---

## One-Command Setup

Run the automated setup script:

```powershell
.\scripts\setup.ps1
```

This installs everything below automatically. See individual sections for manual steps.

---

## Windows Setup (Phase 1 — Verified)

### 1. Install Go

```powershell
winget install --id GoLang.Go --exact --source winget
go version
# Expected: go version go1.26.4 windows/amd64
```

### 2. Install Visual Studio 2022 Build Tools (MSVC)

```powershell
winget install --id Microsoft.VisualStudio.2022.BuildTools --exact --source winget
```

Then install the C++ workload:

```powershell
$bootstrapper = "$env:TEMP\vs_BuildTools.exe"
Invoke-WebRequest -Uri "https://aka.ms/vs/17/release/vs_BuildTools.exe" -OutFile $bootstrapper
$bootstrapper --installPath "C:\Program Files (x86)\Microsoft Visual Studio\2022\BuildTools" `
    --add Microsoft.VisualStudio.Workload.VCTools `
    --add Microsoft.VisualStudio.Component.VC.Tools.x86.x64 `
    --add Microsoft.VisualStudio.Component.Windows11SDK.22621 `
    --quiet --norestart --includeRecommended --wait
```

**Verified**: MSVC 19.44.35228, at `C:\Program Files (x86)\Microsoft Visual Studio\2022\BuildTools\VC\Tools\MSVC\14.44.35207\bin\Hostx64\x64\cl.exe`

### 3. Install CMake & Ninja

```powershell
winget install --id Kitware.CMake --exact --source winget
winget install --id Ninja-build.Ninja --exact --source winget
# Expected: cmake version 4.3.3, ninja 1.13.2
```

### 4. Install Qt6 (via aqtinstall — recommended)

```powershell
pip install aqtinstall
python -m aqt install-qt windows desktop 6.8.2 win64_msvc2022_64 --modules qtmultimedia
```

**Verified**: Qt 6.8.2 installed at `C:\Qt\6.8.2\msvc2022_64`

> **Note**: The Qt installation will be placed at `C:\Qt\6.8.2\msvc2022_64`.  
> Alternative: Use the official Qt Online Installer from https://www.qt.io/download-qt-installer
> and select Qt 6.8.x > MSVC 2022 64-bit > Qt Multimedia.

### 5. Install FFmpeg

```powershell
winget install --id Gyan.FFmpeg --exact --source winget
# Expected: ffmpeg version 8.1.1
# Adds ffmpeg, ffplay, ffprobe to PATH
```

### 6. Install Linting Tools (Optional)

```powershell
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install mvdan.cc/gofumpt@latest
```

---

## Linux Setup (Ubuntu/Debian)

```bash
sudo apt update
sudo apt install -y golang-go cmake ninja-build clang clang-tidy \
    qt6-base-dev qt6-multimedia-dev libavcodec-dev libavformat-dev \
    libavutil-dev libswscale-dev libsqlite3-dev
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install mvdan.cc/gofumpt@latest
```

---

## macOS Setup

```bash
brew install go cmake ninja llvm qt@6 ffmpeg
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install mvdan.cc/gofumpt@latest
```

---

## Cloning & Building

```powershell
git clone https://github.com/your-org/ber.git
cd ber

# Build everything
.\scripts\build.ps1
```

The build script automatically uses the MSVC environment and Qt6 installation.

### Troubleshooting

If `go` is not found when running scripts:

```powershell
$env:Path = [Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [Environment]::GetEnvironmentVariable("Path", "User")
```

---

## Running in Development Mode

### Start the server

```powershell
cd server
go run ./cmd/ber-server --library="C:\path\to\videos" --port=8080

# Or use the pre-built binary:
.\server\ber-server.exe --library="C:\Videos" --scan
```

Open http://localhost:8080 in a browser.

### Start the client

```powershell
cd client
cmake --build build --target ber-client --config Release

# Run from build directory:
.\build\Release\ber-client.exe
```

---

## Useful Commands

```powershell
# Run server tests
cd server
go test -v -count=1 ./...

# Run client tests
$env:Path = "C:\Qt\6.8.2\msvc2022_64\bin;$env:Path"
cd client\build\tests
.\ber-client-tests.exe --verbosity high

# Run all tests via script
.\scripts\test.ps1 -All

# Create release archive
.\scripts\release.ps1

# Format code
gofumpt -l -w server/
clang-format -i client/src/*.cpp client/src/*.h

# Lint
golangci-lint run server/...
clang-tidy client/src/*.cpp -- -std=c++20
```
