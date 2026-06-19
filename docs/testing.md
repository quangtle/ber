# Testing Infrastructure

## Overview

ber uses a three-tier testing strategy:

1. **Unit tests** — Verify individual functions and classes in isolation.
2. **Integration tests** — Verify server ↔ client communication and real FFmpeg/database interactions.
3. **End-to-end tests** — Spin up a real server, start a real client or headless browser, and simulate user flows.

---

## Server Testing (Go)

### Running Tests

```powershell
# All server tests
cd server
go test -v -race -count=1 ./...

# With coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o cover.html
```

### Test Packages

```
server/
├── internal/api/        → handler_test.go      # HTTP endpoint tests (httptest)
├── internal/library/    → library_test.go      # Library scan, add, remove
├── internal/streaming/  → stream_test.go        # Range request logic
├── internal/database/   → db_test.go            # SQLite queries
├── internal/config/     → config_test.go        # Config parsing
└── server_test.go                              # Integration test suite
```

### Conventions

- Use `testing.T` standard library (no third-party test frameworks).
- Table-driven tests for all non-trivial functions.
- Use `testify/assert` and `testify/require` for assertions (added to `go.mod` via `go get github.com/stretchr/testify`).
- Integration tests use build tags: `//go:build integration` to allow skipping during rapid iteration.
- Database tests use an in-memory SQLite instance (`:memory:`).

### Example — Table-Driven Test

```go
func TestLibraryAdd(t *testing.T) {
    tests := []struct {
        name    string
        path    string
        wantErr bool
    }{
        {"valid mp4 file", "testdata/sample.mp4", false},
        {"valid mkv file", "testdata/sample.mkv", false},
        {"nonexistent file", "testdata/missing.mkv", true},
        {"directory path", "testdata/", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            lib := NewLibrary(t.TempDir())
            err := lib.Add(tt.path)
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

---

## Client Testing (C++ / Qt6)

### Framework: Catch2 v3

### Running Tests

```powershell
cd client
cmake --preset test
cmake --build build --target ber-client-tests
ctest --test-dir build --output-on-failure
```

### Test Targets

```
client/tests/
├── CMakeLists.txt           # Catch2 test executable
├── main.cpp                 # Catch2 session
├── test_apiclient.cpp       # Mock HTTP server tests (QNetworkAccessManager)
├── test_librarymodel.cpp    # LibraryModel unit tests
└── test_settings.cpp        # QSettings read/write
```

### Conventions

- Catch2 `TEST_CASE` + `SECTION` for structuring.
- Qt signals/slots tested with `QSignalSpy`.
- Network code tested with `QNetworkAccessManager` + custom `QNetworkReply` mocks or `QTest` + local `QTcpServer`.
- GUI components tested with `QTest::mouseClick`, `QTest::keyClicks`, etc.
- No test should open a real network connection — use `localhost` test servers or mocks.

### Example

```cpp
TEST_CASE("ApiClient parses library response", "[apiclient]") {
    ApiClient client("http://localhost:9999");
    QSignalSpy spy(&client, &ApiClient::libraryLoaded);

    // Inject a mock reply
    auto reply = new QNetworkReply(...);
    // ... setup mock JSON response ...
    client.onLibraryReply(reply);

    REQUIRE(spy.count() == 1);
    auto videos = spy.at(0).at(0).value<QList<VideoInfo>>();
    CHECK(videos.size() == 3);
    CHECK(videos[0].title == "Big Buck Bunny");
}
```

---

## Integration Tests

Located in `tests/integration/`. These tests start a real server instance, add test fixtures, and verify end-to-end behavior.

### Running

```powershell
# Requires FFmpeg test fixtures
.\scripts\build.ps1
.\scripts\test.ps1 -Integration
```

### Test Fixtures

Small synthetic video files located in `tests/fixtures/`:
- `sample_h264.mp4` — 1 second, 640×480, H.264
- `sample_h265.mkv` — 1 second, 640×480, H.265
- `sample_av1.webm` — 1 second, 640×480, AV1
- `test_audio.mp3` — Short audio file

Generate fixtures with:

```powershell
# Requires FFmpeg in PATH
.\scripts\generate-fixtures.ps1
```

---

## End-to-End Tests

Located in `tests/e2e/`. These use Playwright (headless Chromium) to test the web UI against a running server.

### Setup

```powershell
npm install -g @playwright/test
playwright install chromium
```

### Running

```powershell
cd tests/e2e
npx playwright test
```

---

## CI Integration

See `ci/.github/workflows/ci.yml` for the full pipeline. Every push triggers:

1. **Lint** — golangci-lint + clang-tidy
2. **Format** — gofumpt + clang-format check (diff mode)
3. **Build** — Server + client
4. **Unit Tests** — Coverage reports
5. **Integration Tests** — With real FFmpeg

---

## Coverage Targets

| Layer | Current Target | Stretch Target |
|-------|---------------|----------------|
| Server unit | 80% | 90% |
| Client unit | 70% | 85% |
| Integration | 60% | 75% |
| E2E (critical paths) | 100% of defined flows | — |

---

## Test Commands Reference

```powershell
scripts/test.ps1              # Run all unit tests
scripts/test.ps1 -Server      # Server tests only
scripts/test.ps1 -Client      # Client tests only
scripts/test.ps1 -Integration # Integration tests only
scripts/test.ps1 -E2E         # End-to-end tests only
scripts/test.ps1 -Coverage    # With coverage reports
```
