# ber — Local Streaming Software

ber is a lightweight, cross-platform local streaming system that lets you serve a video library from a background server process and watch it from a desktop client or any web browser on your local network.

## Features

- **Efficient**: Built with performance in mind — runs comfortably on a Raspberry Pi.
- **Cross-platform**: Windows, Linux, and macOS (Windows target shipping first).
- **All popular formats**: Leverages FFmpeg for broad format support.
- **Dual client access**: Native desktop client + web browser interface.
- **Zero-config discovery**: Automatic server discovery on LAN (mDNS/Bonjour).
- **Background server**: Runs as a system service / tray daemon.

## Architecture

```
┌──────────────┐     HTTP/WebSocket     ┌────────────────┐
│  ber-client  │ ◄────────────────────── │   ber-server   │
│  (Qt6 C++)   │                        │    (Go HTTP)    │
└──────────────┘                        └───────┬────────┘
                                                │
┌──────────────┐     HTTP/Browser               │
│  Web Browser │ ◄──────────────────────────────┘
│  (any OS)    │                        ┌───────┴────────┐
└──────────────┘                        │  Video Library  │
                                        │  (FFmpeg + fs)  │
                                        └────────────────┘
```

## Quick Start

See [docs/development.md](docs/development.md) to set up your environment and [docs/release.md](docs/release.md) for building distributables.

## Project Structure

```
ber/
├── server/          Go streaming server
├── client/          C++/Qt6 desktop client
├── docs/            Architecture & workflow docs
├── scripts/         Build / test / release helpers
├── ci/              CI/CD configuration
├── tests/           Integration & end-to-end tests
└── third_party/     Vendored dependencies & licenses
```

## License

MIT
