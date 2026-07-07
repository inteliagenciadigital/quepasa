# Configuration

- [Environment Variables Overview](#environment-variables-overview)
- [Cache System Architecture](#cache-system-architecture)
- [Full Variable Reference](#full-variable-reference)

## Environment Variables Overview

Key configuration options (see [docker/.env.example](../../../docker/.env.example) for the complete list):

```bash
# Basic Configuration
DOMAIN=your-domain.com
MASTERKEY=your-secret-key
ACCOUNTSETUP=true  # Enable for first setup

# Database (single shared database: quepasa_* app tables + whatsmeow_* store tables)
# Upgrades from older releases are automatic: legacy tables are renamed and
# standalone whatsmeow.sqlite/quepasa.sqlite files are imported on first start.
# postgres/mysql apply to the whatsmeow store only (app tables stay in local sqlite).
DBDRIVER=sqlite3
DBDATABASE=quepasa
# For postgres/mysql also set DBHOST, DBPORT, DBUSER, DBPASSWORD and DBSSLMODE as needed

# Features
GROUPS=true
READRECEIPTS=true
CALLS=true
WEBSOCKETSSL=false
API_DEFAULT_VERSION=v4  # Unversioned /api/... alias points to v4 or v5

# Performance
CACHELENGTH=800
HISTORYSYNCDAYS=30
```

📖 **[Environment Variables Reference](../../../src/environment/README.md)** — complete configuration documentation.

## Cache System Architecture

QuePasa 3.26+ features a centralized cache system with automatic fallback:

```
┌─────────────────────────────────────────┐
│  Centralized CacheService (Singleton)   │
├─────────────────────────────────────────┤
│                                         │
│  Messages Backend        Queue Backend  │
│  ├─ Memory (sync.Map)    ├─ Memory      │
│  ├─ Disk (JSON)          ├─ Disk        │
│  └─ Redis (go-redis)     └─ Redis       │
│                                         │
│  Auto-Fallback on Failure               │
│  (Enabled when CACHE_INIT_FALLBACK=true)│
└─────────────────────────────────────────┘
```

```bash
# Default: In-memory caching (no setup required)
CACHE_BACKEND=memory

# Persistent disk-based caching
CACHE_BACKEND=disk
CACHE_DISK_PATH=/var/cache/quepasa

# Distributed caching with Redis
CACHE_BACKEND=redis
REDIS_HOST=redis-server
REDIS_PORT=6379
REDIS_PASSWORD=your-password

# Mixed: Memory messages + Redis queue
CACHE_BACKEND=memory
RABBITMQ_CACHE_BACKEND=redis
REDIS_HOST=redis-server
```

Key features:
- Single cache backend for the entire system
- Pluggable backends (memory/disk/redis)
- Automatic fallback to memory on backend failure
- Separate queue backend configuration (optional)
- Environment-based configuration
- Zero external dependencies for the default setup

## Full Variable Reference

### Core Configuration
| Variable | Description | Default |
|----------|-------------|---------|
| `DOMAIN` | Your domain name for the service | `localhost` |
| `WEBAPIPORT` | HTTP server port | `31000` |
| `WEBSOCKETSSL` | Use SSL for WebSocket connections | `false` |
| `MASTERKEY` | Master key for administration | *required* |
| `ACCOUNTSETUP` | Enable account creation setup | `true` |

### WhatsApp Features
| Variable | Description | Default |
|----------|-------------|---------|
| `GROUPS` | Enable group messaging | `true` |
| `BROADCASTS` | Enable broadcast messages | `false` |
| `READRECEIPTS` | Trigger webhooks for read receipts | `false` |
| `CALLS` | Accept incoming calls | `true` |
| `READUPDATE` | Mark chats as read when sending | `true` |

### Cache System (Centralized)

- **memory**: In-process cache (default, no external dependencies)
- **disk**: File-based storage (JSON format)
- **redis**: Distributed cache (for multi-instance deployments)

| Variable | Description | Default | Options |
|----------|-------------|---------|----------|
| `CACHE_BACKEND` | Cache backend type | `memory` | `memory`, `disk`, `redis` |
| `CACHE_DISK_PATH` | Disk storage directory (for disk backend) | `./cache` | *file path* |
| `CACHE_INIT_FALLBACK` | Auto-fallback to memory on backend failure | `true` | `true`, `false` |
| `CACHELENGTH` | Max messages in cache | `800` | *number* |
| `CACHEDAYS` | Days to keep cached messages | `7` | *number* |

**Redis-specific variables** (when `CACHE_BACKEND=redis`):

| Variable | Description | Default |
|----------|-------------|----------|
| `REDIS_HOST` | Redis server hostname | `localhost` |
| `REDIS_PORT` | Redis server port | `6379` |
| `REDIS_USERNAME` | Redis authentication username | `` |
| `REDIS_PASSWORD` | Redis authentication password | `` |
| `REDIS_DATABASE` | Redis database number | `0` |
| `REDIS_KEY_PREFIX` | Prefix for all Redis keys | `quepasa:` |
| `REDIS_POOL_SIZE` | Connection pool size | `10` |
| `REDIS_MAX_RETRIES` | Max reconnection attempts | `3` |
| `REDIS_DIAL_TIMEOUT` | Connection timeout (seconds) | `5` |
| `REDIS_READ_TIMEOUT` | Read operation timeout (seconds) | `3` |
| `REDIS_WRITE_TIMEOUT` | Write operation timeout (seconds) | `3` |

**RabbitMQ Queue Backend** (independent from message cache):

| Variable | Description | Default |
|----------|-------------|----------|
| `RABBITMQ_CACHE_BACKEND` | Queue backend (disk/redis/memory) | *inherits CACHE_BACKEND* |
| `RABBITMQ_CACHE_DISK_PATH` | Queue disk storage path | *inherits CACHE_DISK_PATH* |
| `RABBITMQ_CACHE_QUEUE_KEY` | Redis queue namespace | `rabbitmq_retry` |
| `RABBITMQ_CACHELENGTH` | Max messages in retry queue (legacy) | `100000` |

### Performance & Sync

| Variable | Description | Default |
|----------|-------------|----------|
| `HISTORYSYNCDAYS` | Days of history to sync on QR scan | `30` |
| `SYNOPSISLENGTH` | Length for message synopsis | `50` |

### Database Configuration

These `DB*` variables configure the **Whatsmeow SQL store** used during startup.
They do **not** currently move the internal QuePasa application database, which
still follows the local `quepasa.sqlite` / `quepasa.db` path in the current code.

| Variable | Description | Default |
|----------|-------------|---------|
| `DBDRIVER` | SQL driver for the Whatsmeow store (`sqlite3`, `postgres`, `mysql`) | `sqlite3` |
| `DBHOST` | Host for `postgres` / `mysql`; ignored by `sqlite3` | empty |
| `DBPORT` | Port for `postgres` / `mysql`; ignored by `sqlite3` | empty |
| `DBDATABASE` | Database name for `postgres` / `mysql`, or sqlite base file path/name for the Whatsmeow store | runtime fallback: `whatsmeow` when empty with `sqlite3` |
| `DBUSER` | User for `postgres` / `mysql`; ignored by `sqlite3` | empty |
| `DBPASSWORD` | Password for `postgres` / `mysql`; ignored by `sqlite3` | empty |
| `DBSSLMODE` | PostgreSQL `sslmode`; usually unused by `sqlite3` / `mysql` | empty |

### Logging & Debug
| Variable | Description | Options |
|----------|-------------|---------|
| `LOGLEVEL` | Application log level | `ERROR`, `WARN`, `INFO`, `DEBUG`, `TRACE` |
| `WHATSMEOW_LOGLEVEL` | WhatsApp library log level | `error`, `warn`, `info`, `debug` |
| `HTTPLOGS` | Log HTTP requests | `true`, `false` |
| `DEBUGREQUESTS` | Debug API requests | `true`, `false` |

### Media & Conversion
| Variable | Description | Default |
|----------|-------------|---------|
| `CONVERT_PNG_TO_JPG` | Convert PNG to JPG format | `false` |
| `COMPATIBLE_MIME_AS_AUDIO` | Convert audio to OGG/PTT | `true` |
| `REMOVEDIGIT9` | Remove digit 9 from BR numbers | `false` |

### Regional Settings
| Variable | Description | Default |
|----------|-------------|---------|
| `TZ` | Timezone | `America/Sao_Paulo` |
| `APP_TITLE` | App title suffix | `QuePasa` |
| `PRESENCE` | Default presence state | `unavailable` |
