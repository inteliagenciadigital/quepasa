# API

- [Core Endpoints](#core-endpoints)
- [API Versions](#api-versions)
- [Examples](#examples)
- [Authentication](#authentication)

[![Run in Postman](https://run.pstmn.io/button.svg)](https://god.gw.postman.com/run-collection/5047984-405506cf-59f5-479e-b512-4ba5b935411b?action=collection%2Ffork&source=rip_markdown&collection-url=entityId%3D5047984-405506cf-59f5-479e-b512-4ba5b935411b%26entityType%3Dcollection%26workspaceId%3Dbd72aaba-0c31-40ad-801c-d5ba19184aff#?env%5BQuepasa%5D=W3sia2V5IjoiYmFzZVVybCIsInZhbHVlIjoiIiwiZW5hYmxlZCI6dHJ1ZSwidHlwZSI6ImRlZmF1bHQiLCJzZXNzaW9uVmFsdWUiOiIiLCJjb21wbGV0ZVNlc3Npb25WYWx1ZSI6IiIsInNlc3Npb25JbmRleCI6MH0seyJrZXkiOiJ0b2tlbiIsInZhbHVlIjoiIiwiZW5hYmxlZCI6dHJ1ZSwidHlwZSI6ImRlZmF1bHQiLCJzZXNzaW9uVmFsdWUiOiIiLCJjb21wbGV0ZVNlc3Npb25WYWx1ZSI6IiIsInNlc3Npb25JbmRleCI6MX0seyJrZXkiOiJjaGF0SWQiLCJ2YWx1ZSI6IiIsImVuYWJsZWQiOnRydWUsInR5cGUiOiJkZWZhdWx0Iiwic2Vzc2lvblZhbHVlIjoiIiwiY29tcGxldGVTZXNzaW9uVmFsdWUiOiIiLCJzZXNzaW9uSW5kZXgiOjJ9LHsia2V5IjoiZmlsZU5hbWUiLCJ2YWx1ZSI6IiIsImVuYWJsZWQiOnRydWUsInR5cGUiOiJkZWZhdWx0Iiwic2Vzc2lvblZhbHVlIjoiIiwiY29tcGxldGVTZXNzaW9uVmFsdWUiOiIiLCJzZXNzaW9uSW5kZXgiOjN9LHsia2V5IjoidGV4dCIsInZhbHVlIjoiIiwiZW5hYmxlZCI6dHJ1ZSwidHlwZSI6ImRlZmF1bHQiLCJzZXNzaW9uVmFsdWUiOiIiLCJjb21wbGV0ZVNlc3Npb25WYWx1ZSI6IiIsInNlc3Npb25JbmRleCI6NH0seyJrZXkiOiJ0cmFja0lkIiwidmFsdWUiOiJwb3N0bWFuIiwiZW5hYmxlZCI6dHJ1ZSwidHlwZSI6ImRlZmF1bHQiLCJzZXNzaW9uVmFsdWUiOiJwb3N0bWFuIiwiY29tcGxldGVTZXNzaW9uVmFsdWUiOiJwb3N0bWFuIiwic2Vzc2lvbkluZGV4Ijo1fV0=)

## Core Endpoints
- **Messages**: `/send`
- **Media**: `/send`
- **Groups**: `/groups/`
- **Webhooks**: `/webhook`
- **RabbitMQ**: `/rabbitmq`

## API Versions
- **v4** (Latest) - Recommended for new integrations
- **v5** - Canonical, what the official Vue.js SPA uses internally
- **v3** - Legacy support
- **v2** - Legacy support
- **v1** - Deprecated

The unversioned alias under your configured API prefix follows `API_DEFAULT_VERSION` (see [Configuration](../configuration/README.md)):
- `/api/v4/...` → always legacy v4
- `/api/v5/...` → always canonical v5
- `/api/...` → whichever version is selected by `API_DEFAULT_VERSION`

## Examples

```bash
# Connect and get QR code
# token could be empty, if empty a new token will be generated
# user is the user that will be manage this connection

curl --location 'localhost:31000/scan' \
  --header 'Accept: application/json' \
  --header 'X-QUEPASA-USER: :user' \
  --header 'X-QUEPASA-TOKEN: :token' \
  --data ''

# Send a message
curl --location 'localhost:31000/send' \
  --header 'Accept: application/json' \
  --header 'X-QUEPASA-TRACKID: :trackid' \
  --header 'X-QUEPASA-CHATID: :chatid' \
  --header 'Content-Type: application/json' \
  --header 'X-QUEPASA-TOKEN: :token' \
  --data '{
      "text": "Hello World ! \nHello World !"
  }'

# Set webhook
curl --location 'localhost:31000/webhook' \
  --header 'Accept: application/json' \
  --header 'Content-Type: application/json' \
  --header 'X-QUEPASA-TOKEN: :token' \
  --data '{
      "url": "https://webhook.example.com/webhook/5465465241654",
      "forwardinternal": true,
      "trackid": "custom-track",
      "extra": {
        "clientId": "12345",
        "company": "myCompany",
        "enviroment": "production",
        "version": "1.0"
      }
  }'
```

## Authentication

QuePasa supports four authentication modes for secure API access:

1. **JWT (User-Level Access)** — Full access to all user sessions
   - Use the `Authorization: Bearer <jwt-token>` header
   - Recommended for multi-session applications
   - Managed by `SIGNING_SECRET` environment variable

2. **X-QUEPASA-TOKEN (Session-Scoped Access)** — Access to a single session
   - Use the `X-QUEPASA-TOKEN: <token>` header
   - Recommended for single-session applications
   - Automatically created when `RELAXED_SESSIONS=true` (default)
   - Restricted to authenticated session only (cannot access other sessions)

3. **X-QUEPASA-MASTERKEY** — Administration and setup
   - Use the `X-QUEPASA-MASTERKEY: <masterkey>` header
   - Required for setup operations when `RELAXED_SESSIONS=false`
   - Managed by `MASTERKEY` environment variable

4. **Anonymous** — Public endpoints only
   - No authentication required for health checks, version info, etc.
   - No headers needed

```bash
# JWT authentication (user-level)
curl -H "Authorization: Bearer your-jwt-token" \
  http://localhost:31000/api/servers

# Session token authentication (single session)
curl -H "X-QUEPASA-TOKEN: your-session-token" \
  http://localhost:31000/api/servers

# Master key (administration)
curl -H "X-QUEPASA-MASTERKEY: your-masterkey" \
  http://localhost:31000/api/sessions

# Anonymous (public endpoints)
curl http://localhost:31000/health
```

📖 **[Complete Authentication Guide](../../USAGE-authentication-modes.md)** — all modes, use cases, security considerations, advanced examples.

📖 **[Environment Discovery Guide](../../USAGE-environment-discovery.md)** — query server capabilities with `GET /api/system/environment`.

## Connection States

QuePasa exposes connection states such as `Ready`, `Stopped`, `Disconnected`, and `Failed` to represent the runtime status of each WhatsApp server.

📖 **[Connection States Guide](../../CONNECTION_STATES.md)** — detailed explanation of each state, health semantics, and which states are currently emitted by the runtime.
