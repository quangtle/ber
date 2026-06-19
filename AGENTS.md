# Agent Instructions for ber

## Requirements

- **Language**: Server = Go 1.22+, Client = C++20 with Qt6.
- **Testing**: `go test ./...` for server; `ctest` for client.
- **Linting**: `golangci-lint run` for Go; `clang-tidy` for C++.
- **Formatting**: `gofumpt -l -w` for Go; `clang-format` for C++.
- **Build**: `scripts/build.ps1` on Windows, `scripts/build.sh` on Unix.

## Conventions

- No explanatory comments in source code unless the logic is non-obvious.
- Keep functions small and focused.
- Use idiomatic Go (no ORMs, prefer `database/sql` with raw queries).
- In C++ use RAII, avoid raw pointers, prefer `std::unique_ptr`.
- Server API responses use JSON with a consistent envelope: `{"ok": true, "data": ...}` or `{"ok": false, "error": "..."}`.
