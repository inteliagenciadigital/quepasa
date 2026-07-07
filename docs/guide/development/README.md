# Development & Contributing

- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Building](#building)
- [Contributing](#contributing)

## Architecture

QuePasa is built with:
- **Backend**: Go with [Whatsmeow](https://github.com/tulir/whatsmeow) library
- **Storage**: internal QuePasa data uses local SQLite by default, while the Whatsmeow store is configurable through `DB*` variables (`sqlite3`, `postgres`, or `mysql`)
- **API**: RESTful HTTP endpoints
- **Real-time**: WebSocket support for live updates

For deeper architecture notes (ADRs, package maps, roadmap), see the [`docs/`](../..) folder — start at [ARCHITECTURE-INDEX.md](../../ARCHITECTURE-INDEX.md).

## Project Structure
```
├── src/                    # Go source code
├── docker/                 # Docker configuration
├── extra/                  # Integration examples
│   ├── chatwoot/          # Chatwoot integration
│   ├── n8n+chatwoot/      # N8N workflow examples
│   └── typebot/           # TypeBot integration
├── docs/                   # Documentation (architecture, ADRs, usage guides)
│   └── guide/              # User-facing guides (this section)
└── helpers/                # Installation helpers
```

## Building
```bash
# Development build
go build -o .dist/quepasa-dev src/main.go

# Production build
go build -ldflags="-s -w" -o .dist/quepasa-prod src/main.go
```

Multi-platform release binaries (linux/windows/darwin, amd64/arm64/386) are built and attached automatically to each GitHub Release by [`.github/workflows/docker.yml`](../../../.github/workflows/docker.yml) whenever `QpVersion` changes on `main`. To rebuild binaries for an existing tag manually, use the [`Release` workflow](../../../.github/workflows/release.yml) (`workflow_dispatch`).

## Contributing

Issues and pull requests are welcome on [GitHub](https://github.com/nocodeleaks/quepasa-) or via the [Radicle mirror](../installation/README.md#p2p-mirror-radicle).

- **Security**: This application has not been security audited. Use at your own risk.
- **Unofficial**: This is a third-party project, not affiliated with WhatsApp.
- **Terms**: Ensure compliance with WhatsApp's Terms of Service.
- **Rate Limits**: Respect WhatsApp's rate limiting to avoid account suspension.
