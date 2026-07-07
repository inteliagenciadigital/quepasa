<!-- VERSION: 5.26.0702.1713 -->
<p align="center">
	<img src="src/assets/favicon.png" alt="Quepasa-logo" width="100" />
</p>

<h1 align="center">QuePasa</h1>

<p align="center">
	<strong>An open-source, free-license micro web-application to exchange messages with the WhatsApp Platform.</strong>
</p>

<p align="center">
	<a href="https://github.com/nocodeleaks/quepasa-/actions/workflows/go.yml"><img src="https://github.com/nocodeleaks/quepasa-/actions/workflows/go.yml/badge.svg" alt="Go Build" /></a>
	<a href="https://github.com/nocodeleaks/quepasa-/actions/workflows/docker.yml"><img src="https://github.com/nocodeleaks/quepasa-/actions/workflows/docker.yml/badge.svg" alt="Docker Publish" /></a>
	<a href="https://github.com/nocodeleaks/quepasa-/releases/latest"><img src="https://img.shields.io/github/v/release/nocodeleaks/quepasa-" alt="Latest Release" /></a>
	<a href="https://hub.docker.com/r/codeleaks/quepasa"><img src="https://img.shields.io/docker/pulls/codeleaks/quepasa" alt="Docker Pulls" /></a>
	<a href="LICENSE.md"><img src="https://img.shields.io/badge/License-AGPL%203.0-lightgrey.svg" alt="License GNU AGPL v3.0" /></a>
</p>

<p align="center">
	<a href="https://run.pstmn.io/button.svg"><img src="https://run.pstmn.io/button.svg" alt="Run in Postman" /></a>
</p>

<hr />

## Quick Start

```bash
git clone https://github.com/nocodeleaks/quepasa-.git
cd quepasa-/docker
cp .env.example .env
docker-compose up -d --build
```

Web interface: `http://localhost:31000` · Docker image: **[codeleaks/quepasa](https://hub.docker.com/r/codeleaks/quepasa)**

Also mirrored P2P on **[Radicle](https://radicle.xyz)** — no account needed: `rad clone rad:z4N4hVxiroGtMpgB5AyzvfcptdfgD`

## Features

QuePasa provides a simple HTTP API to integrate WhatsApp messaging into your applications:

- 📱 QR Code authentication — easy WhatsApp Web connection setup
- 💾 Persistent sessions — account data and keys stored securely
- 🔗 HTTP API for sending messages, media & documents, receiving via webhooks, downloading attachments, managing contacts/groups
- 🔄 Webhook support — real-time message notifications
- 📊 Message history sync — configurable retrieval window
- 🎯 Read receipts, message reactions, broadcasts, call handling, presence management

## Documentation

| Guide | Covers |
|---|---|
| **[Installation](docs/guide/installation/README.md)** | Docker, local build from source, Radicle mirror, Swagger generation |
| **[API](docs/guide/api/README.md)** | Endpoints, versions, curl examples, authentication modes, connection states |
| **[Configuration](docs/guide/configuration/README.md)** | Environment variables, cache backends (memory/disk/redis) |
| **[Integrations](docs/guide/integrations/README.md)** | N8N, Chatwoot, TypeBot |
| **[Development](docs/guide/development/README.md)** | Architecture, project structure, building, contributing |
| **[Community & Support](docs/guide/community/README.md)** | Telegram, issues, license, references |

Deeper architecture notes (ADRs, package maps, roadmap) live under [`docs/`](docs/), starting at [ARCHITECTURE-INDEX.md](docs/ARCHITECTURE-INDEX.md).

## Important Notices

- **Security**: this application has not been security audited. Use at your own risk.
- **Unofficial**: third-party project, not affiliated with WhatsApp.
- **Terms**: ensure compliance with WhatsApp's Terms of Service.

## License

QuePasa is free software licensed under **AGPL-3.0** — see [Community & Support](docs/guide/community/README.md#license) for details.

---

<p align="center">
	<img src="https://telegram.org/favicon.ico" alt="Telegram-logo" width="20" />
	<a href="https://t.me/quepasa_api">Telegram Group</a> ·
	<a href="https://t.me/quepasa_channel">Telegram Channel</a>
</p>
<p align="center">
	<sub>Logo by <a href="https://agenciaoctos.com.br">Lukas Prais</a> · Made with ❤️ by the QuePasa Community</sub>
</p>
