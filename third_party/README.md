# Third-Party Dependencies

## Server (Go Modules)

All Go dependencies are managed via `go.mod` / `go.sum` in `server/`.  
Key dependencies:

- `github.com/go-chi/chi/v5` — HTTP router (BSD-3)
- `modernc.org/sqlite` — Pure-Go SQLite (BSD-3)
- `github.com/google/uuid` — UUID generation (BSD-3)
- `github.com/stretchr/testify` — Test assertions (MIT)

## Client (C++/Qt6)

- **Qt 6.6+** — Cross-platform GUI framework (LGPLv3 / GPLv3)
- **Catch2 v3** — C++ unit testing framework (BSL-1.0)
- **FFmpeg** — Video decoding / probing (LGPLv2.1+)

## Licenses

See individual package documentation for full license text.
