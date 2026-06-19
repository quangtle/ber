# Software Architecture

## Overview

ber consists of two main components: a **streaming server** (`ber-server`) written in Go, and a **desktop client** (`ber-client`) written in C++ with Qt6. The server also serves a web-based UI accessible from any modern browser on the local network.

## Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│                     Local Network                        │
│                                                         │
│  ┌──────────────┐   HTTP/REST + WebSocket   ┌─────────┐│
│  │  ber-client  │◄─────────────────────────►│         ││
│  │  (Qt6 C++)   │                           │  ber-   ││
│  └──────────────┘                           │  server ││
│                                              │  (Go)   ││
│  ┌──────────────┐   HTTP (browser)          │         ││
│  │  Web Browser │◄─────────────────────────►│         ││
│  │  (any OS)    │                           └────┬────┘│
│  └──────────────┘                                │     │
│                                                  │     │
│                                         ┌────────┴────┐│
│                                         │  File System ││
│                                         │  + SQLite DB ││
│                                         └─────────────┘│
└─────────────────────────────────────────────────────────┘
```

## Technology Choices

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| Server | Go 1.22+ | Excellent concurrency model, fast compilation, single-binary deploy, trivial cross-compilation |
| Client | C++20 + Qt6 | Mature cross-platform GUI toolkit, hardware-accelerated video, native look and feel |
| Web UI | Vanilla JS + CSS | Zero dependency overhead; served directly from the server binary via `embed` |
| Video Engine | FFmpeg (libav*) | Universal format support, hardware acceleration on all platforms |
| Database | SQLite (via `modernc.org/sqlite`) | Embedded, zero-admin, single-file, perfect for a home media server |
| Service Discovery | mDNS (Bonjour) | Zero-configuration LAN discovery; no manual IP entry needed |
| Streaming | HTTP range requests + WebSocket seeking | Stateless, cache-friendly, works through proxies, resume support |

## Server Architecture (`server/`)

### Internal Packages

```
server/
├── cmd/ber-server/main.go        # Entry point: flags, config, startup
├── internal/
│   ├── api/                       # HTTP handlers (REST endpoints)
│   │   └── handler.go
│   ├── config/                    # Configuration loading (flags, env, file)
│   │   └── config.go
│   ├── database/                  # SQLite schema and queries
│   │   └── db.go
│   ├── ffmpeg/                    # FFmpeg probe and transcode wrapper
│   │   └── probe.go
│   ├── library/                   # Video library scanning and management
│   │   └── library.go
│   ├── mdns/                      # mDNS service advertisement
│   │   └── mdns.go
│   └── streaming/                 # HTTP range-based video streaming
│       └── stream.go
└── cmd/ber-server/web/            # Embedded web UI (go:embed)
    ├── index.html
    ├── css/style.css
    └── js/app.js
```

### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/status` | Server health + library stats |
| `GET` | `/api/library` | List all videos in library |
| `GET` | `/api/library/:id` | Get single video metadata |
| `POST` | `/api/library/scan` | Trigger library rescan |
| `POST` | `/api/library/add` | Add a file or directory |
| `DELETE` | `/api/library/:id` | Remove from library |
| `GET` | `/api/stream/:id` | Stream video (HTTP range requests) |
| `GET` | `/api/stream/:id/thumbnail` | Get video thumbnail |
| `GET` | `/api/stream/:id/subtitles` | Get subtitle tracks |
| `WS` | `/api/ws` | WebSocket (real-time events) |
| `GET` | `/*` | Static file server for web UI |

### Data Flow — Video Streaming

```
Client                          Server                         Filesystem
  │                               │                               │
  │  GET /api/stream/:id          │                               │
  │  Range: bytes=0-              │                               │
  │──────────────────────────────►│                               │
  │                               │  Lookup metadata in SQLite    │
  │                               │  Open file handle             │
  │                               │  Seek to requested byte range │
  │                               │──────────────────────────────►│
  │                               │◄──────────────────────────────│
  │  206 Partial Content          │                               │
  │  Content-Range: bytes 0-999   │                               │
  │◄──────────────────────────────│                               │
  │  [binary stream data]         │                               │
  │◄──────────────────────────────│                               │
```

### Database Schema (SQLite)

```sql
CREATE TABLE videos (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    file_path   TEXT NOT NULL UNIQUE,
    file_size   INTEGER NOT NULL,
    duration    REAL,
    width       INTEGER,
    height      INTEGER,
    codec       TEXT,
    bitrate     INTEGER,
    container   TEXT,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_videos_title ON videos(title);
CREATE INDEX idx_videos_path  ON videos(file_path);
```

## Client Architecture (`client/`)

```
client/
├── CMakeLists.txt                # Top-level CMake build
├── src/
│   ├── main.cpp                  # Entry point + QApplication setup
│   ├── mainwindow.h/.cpp         # Main window shell (menus, toolbar, status)
│   ├── connectiondialog.h/.cpp   # Server discovery / manual IP entry
│   ├── libraryview.h/.cpp        # Video grid / list widget
│   ├── playerwidget.h/.cpp       # QMainWindow child wrapping the video surface
│   ├── videoplayer.h/.cpp        # Qt Multimedia / libVLC playback backend
│   ├── apiclient.h/.cpp          # HTTP client for server REST API
│   └── settings.h/.cpp           # QSettings wrapper (remember server, prefs)
├── tests/                        # Catch2 unit tests
│   ├── CMakeLists.txt
│   ├── main.cpp
│   ├── test_apiclient.cpp
│   └── test_settings.cpp
└── resources/
    ├── icons/
    └── ber.qrc                   # Qt resource file
```

### Client Class Hierarchy

```
QApplication
 └── MainWindow (QMainWindow)
      ├── MenuBar (File, View, Help)
      ├── ToolBar (Play, Pause, Volume, Fullscreen)
      ├── LibraryView (QListView / QGridView)
      │    └── Model: LibraryModel (QAbstractListModel)
      ├── PlayerWidget (QWidget)
      │    └── VideoPlayer (QMediaPlayer / VLC::MediaPlayer)
      ├── StatusBar
      └── ConnectionDialog (QDialog) [modal on startup]
```

## Web UI

The web UI is a single-page application built with vanilla JavaScript. It communicates with the server via the REST API and uses the HTML5 `<video>` element for playback. No build step or framework is required — the files are embedded into the Go server binary at compile time via `//go:embed`.

## Performance Goals

- **Server memory**: < 50 MB idle (no active streams), < 200 MB with 3 concurrent transcodes
- **Client memory**: < 100 MB idle, < 300 MB during 4K playback
- **CPU**: No transcoding for direct-play formats; on-demand transcoding only when format is unsupported by client
- **Startup time**: Server < 500 ms, Client < 2 s

## Cross-Platform Strategy

| Platform | Server | Client | Status |
|----------|--------|--------|--------|
| Windows | Go cross-compile | MSVC 2022 + Qt6 | **Phase 1** |
| Linux | Native / cross | GCC + Qt6 | Phase 2 |
| macOS | Native / cross | Clang + Qt6 | Phase 3 |
| Raspberry Pi OS | ARM64 cross | (Web UI only) | Phase 2 |

## Security

- No authentication in v1 (local network only; bind to `127.0.0.1` or LAN interface)
- Path traversal protection in all file-serving endpoints
- Input validation on all API endpoints
- CORS restricted to same-origin and known client User-Agent strings
