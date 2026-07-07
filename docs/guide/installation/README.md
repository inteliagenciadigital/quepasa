# Installation

- [Docker (recommended)](#docker-recommended)
- [Local Development](#local-development)
- [P2P Mirror (Radicle)](#p2p-mirror-radicle)
- [Swagger / API Docs Generation](#swagger--api-docs-generation)

## Docker (recommended)

The fastest way to get QuePasa running.

```bash
# Clone the repository
git clone https://github.com/nocodeleaks/quepasa-.git
cd quepasa-/docker

# Configure environment
cp .env.example .env
# Edit .env with your settings

# Start with Docker Compose
docker-compose up -d --build
```

Web interface: `http://localhost:31000`

Images are published to Docker Hub: **[codeleaks/quepasa](https://hub.docker.com/r/codeleaks/quepasa)**

```bash
docker pull codeleaks/quepasa:latest
```

Tags:
- `latest` — last build from `main`
- `<version>` (e.g. `5.26.0702.1713`) — pinned release version, built from `main`
- `dev-latest` / `dev-<version>` — builds from `develop`, for testing unreleased changes

📖 **[Complete Docker Setup Guide](../../../docker/docker.md)** — full configuration options, `.env` reference, troubleshooting.

## Local Development

For development or custom installations.

### Prerequisites
- **Go 1.20+** — [Download here](https://golang.org/dl/)
- **PostgreSQL** (optional) — only needed if you point the Whatsmeow store at `postgres`/`mysql`
- **Git**

### Build from Source
```bash
git clone https://github.com/nocodeleaks/quepasa-.git
cd quepasa-/src

go mod download
go build -o quepasa main.go
./quepasa
```

See **[Development Guide](../development/README.md)** for project structure, build variants and contribution notes.

## P2P Mirror (Radicle)

This repository is also seeded on [Radicle](https://radicle.xyz), a peer-to-peer alternative to GitHub. No account needed, just the `rad` CLI:

```bash
# Install the Radicle CLI (see https://radicle.xyz for details)
curl -sSf https://radicle.xyz/install | sh

# Create your local identity (one-time)
rad auth

# Clone quepasa's `develop` branch from the P2P network
rad clone rad:z4N4hVxiroGtMpgB5AyzvfcptdfgD
```

Repository ID (RID): `rad:z4N4hVxiroGtMpgB5AyzvfcptdfgD` — only `develop` is published there.

## Swagger / API Docs Generation

QuePasa uses Swagger/OpenAPI for API documentation.

```bash
# Install swag CLI tool (one-time setup)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate/update API documentation
cd src
swag init --output ./swagger

# Or use the provided script
# Windows: double-click generate-swagger.bat
# Or run: .\generate-swagger.bat

# Or use VS Code task: Ctrl+Shift+P → "Tasks: Run Task" → "Generate Swagger Docs"
```

Docs are served at `http://localhost:PORT/swagger` (with or without trailing slash) while the application is running.
