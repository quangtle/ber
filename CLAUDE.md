# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```powershell
# Server
cd server; go test -v -count=1 ./...          # all server tests
go test -v -race -coverprofile=coverage.out ./...
go build -o ber-server.exe .\cmd\ber-server\   # build server binary
go build -o ber-desktop.exe .\cmd\ber-desktop\ # build desktop (Edge wrapper)
go run .\cmd\ber-server\ --library="C:\Videos" --scan  # run dev server

# Client (C++/Qt6)
cd client
cmake -S . -B build -G Ninja -DCMAKE_BUILD_TYPE=Debug -DCMAKE_PREFIX_PATH="C:\Qt\6.8.2\msvc2022_64"
cmake --build build --config Debug
cmake --build build --target ber-client-tests --config Debug
build\tests\ber-client-tests.exe --verbosity high
ctest --test-dir build --output-on-failure

# Scripts
.\scripts\test.ps1 -Server     # server tests
.\scripts\test.ps1 -All        # all tests
.\scripts\build.ps1            # build server + client
.\scripts\release.ps1          # build + pack distribution
```

- **No external Go dependencies** — the server has a `go.mod` with only the module path and `go 1.22`.
- C++ client uses Catch2 v3 (auto-downloaded via CMake FetchContent).

## Architecture

### Two components + web UI

**ber-server** (Go 1.22+, `server/cmd/ber-server/main.go`) — HTTP streaming server with:
- JSON-file-backed "database" (`server/internal/database/db.go`) — stores video records as a JSON array. Thread-safe via `sync.RWMutex`. No SQLite or any external dep.
- Library scanner (`server/internal/library/library.go`) — walks a directory, registers supported video files (`.mp4`, `.mkv`, `.avi`, `.mov`, `.webm`, etc.) by extension. Generates v4 UUIDs via `crypto/rand`. No FFmpeg dependency.
- HTTP router + handlers (`server/internal/api/handler.go`) — standard `net/http` with Go 1.22+ pattern matching (`GET /api/library/{id}`). JSON envelope: `{"ok": true, "data": ...}`.
- Streaming uses `http.ServeContent` — handles range requests, content-type sniffing, and caching headers automatically.
- Windows system tray (`server/internal/tray/tray_windows.go`) — raw Win32 API via `syscall`. Hidden window + `NOTIFYICONDATA`. Fallback to signal-based blocking on other platforms (`tray_other.go`).
- Client auto-discovery: UDP beacon broadcasting on `255.255.255.255:10001` every 2s.

**ber-client** (C++20 + Qt6, `client/`) — desktop GUI app:
- Qt6 Multimedia for video playback (`videoplayer.h/.cpp`)
- REST API client (`apiclient.h/.cpp`) using `QNetworkAccessManager`
- Library grid/list view (`libraryview.h/.cpp`)
- Server connection dialog with discovery (`connectiondialog.h/.cpp`)
- QSettings persistence (`settings.h/.cpp`)
- CMake + Ninja build, Catch2 unit tests

**ber-desktop** (Go, `server/cmd/ber-desktop/main.go`) — lightweight wrapper that starts ber-server and opens Microsoft Edge in `--app` mode for a desktop-app-like experience.

**Web UI** (`server/cmd/ber-server/web/`) — vanilla HTML/CSS/JS embedded in the server binary via `//go:embed`.

### Routes

| Method | Path | Handler |
|--------|------|---------|
| GET | `/api/status` | Server health + library stats |
| GET | `/api/library` | List all videos |
| GET | `/api/library/{id}` | Single video metadata |
| POST | `/api/library/scan` | Scan configured library path |
| POST | `/api/library/add` | Add a file path to library |
| POST | `/api/library/scan-dir` | Scan an arbitrary directory |
| DELETE | `/api/library/{id}` | Remove a video |
| GET | `/api/stream/{id}` | Stream video file |
| GET | `/api/stream/{id}/thumbnail` | Thumbnail image |
| GET | `/api/browse` | Browse filesystem directories |
| /* | Static file server for web UI |

### Key conventions

- No explanatory comments in source code unless logic is non-obvious.
- Many `// ponytail:` comments mark deliberate simplifications (e.g., global lock, fallback behavior, Windows-only shortcuts).
- Go: standard library only (except test imports `testing` — no testify). Idiomatic `net/http`, no frameworks.
- C++: RAII, `std::unique_ptr`, no raw owning pointers.
- Tests: standard `testing.T` with table-driven style. Build tag `//go:build integration` for integration tests.
