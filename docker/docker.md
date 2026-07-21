# QuePasa Docker Installation Guide

## English Version

### Prerequisites
- Docker and Docker Compose installed
- Basic knowledge of environment variables
- Access to configure your domain/network

### Installation Steps

1. **Clone or download the project**
   ```bash
   git clone https://github.com/inteliagenciadigital/quepasa.git
   cd quepasa
   ```

2. **Configure environment variables**
   ```bash
   cd docker
   cp .env.example .env
   # Edit .env file with your specific configurations
   ```

3. **Edit the `.env` file with your settings**
   - **DOMAIN**: Set your domain (e.g., `quepasa.yourdomain.com`)
   - **EMAIL**: Set your administrator email
   - **MASTERKEY**: Change the default master key for security
   - **CORS_ALLOWED_ORIGINS**: Comma-separated browser origins allowed cross-origin (e.g. `https://app.example,https://admin.example`). Leave empty for same-origin only. Use `*` only for trusted/dev setups (disables credentials).
   - **RELAXED_SESSIONS**: `true` (default) lets any authenticated user create sessions without the master key. Set `false` for stricter multi-tenant control.
   - **PASSWORD**: Set a strong password
   - **DBDRIVER** / **DBDATABASE**: Define the **single shared database** (quepasa_* app tables + whatsmeow_* store tables)
   - **DBPASSWORD**: Set a secure password only when using PostgreSQL/MySQL (`DBDRIVER=postgres` or `mysql`)
   - **SIGNING_SECRET**: Change the default signing secret
   - **WEBSOCKETSSL**: Set to `true` if using HTTPS/SSL
   - **LOGLEVEL**: Adjust logging level (ERROR, WARN, INFO, DEBUG, TRACE)
   - **TZ**: Set your timezone
   - **OAuth (optional)**: Enable external authentication via OIDC provider (e.g. Keycloak, Auth0, Google):
     - `OAUTH_ENABLED=true`
     - `OAUTH_PROVIDER_URL=https://identity.example.com`
     - `OAUTH_CLIENT_ID=quepasa-client`
     - `OAUTH_CLIENT_SECRET=<secret>`
     - `OAUTH_REDIRECT_URI=https://quepasa.yourdomain.com/oauth/callback`
     - `OAUTH_SCOPES=openid,email,profile` (optional, defaults to standard OIDC)
     - See `docs/USAGE-oauth-authentication.md` for details.

4. **Optional: Review docker-compose.yml**
   - The compose file now uses environment variables from `.env`
   - Includes a PostgreSQL database service for the Whatsmeow store
   - Keep the `DB*` variables aligned with the compose database service when `DBDRIVER=postgres`
   - The internal QuePasa application DB is still local in the current code path

5. **Build and run the container**
   ```bash
   # Option 1: Build and run in one command
   docker-compose up -d --build
   
   # Option 2: Build first, then run
   docker-compose build
   docker-compose up -d
   
   # Using newer Docker Compose syntax
   docker compose up -d --build
   ```

6. **Verify installation**
   ```bash
   # Check container status
   docker-compose ps
   
   # Check logs
   docker-compose logs -f quepasa
   ```

### Important Configuration Notes

- **Environment File**: All configurations are now in `.env` file for better management
- **Database**: PostgreSQL service included with automatic setup
- **First Setup**: Set `ACCOUNTSETUP=true` for initial configuration
- **Security**: Change all default passwords and secrets in `.env`
- **SSL**: Set `WEBSOCKETSSL=true` if using HTTPS
- **Network**: Uses internal Docker network `quepasa_network`
- **Ports**: Default port 31000, configurable via `QUEPASA_EXTERNAL_PORT`

### Environment Variables Overview

The `.env` file contains all necessary configurations organized in sections:
- **Basic Config**: Domain, setup flags, master key
- **Authentication**: Email, passwords, auth settings  
- **WhatsApp Features**: Groups, broadcasts, calls, receipts
- **Logging**: Log levels for application and WhatsApp
- **Database**: PostgreSQL connection settings
- **Performance**: Cache, memory, sync settings
- **Debug**: Various debugging options

---

## VersÃ£o em PortuguÃªs

### PrÃ©-requisitos
- Docker e Docker Compose instalados
- Conhecimento bÃ¡sico de variÃ¡veis de ambiente
- Acesso para configurar seu domÃ­nio/rede

### Passos de InstalaÃ§Ã£o

1. **Clone ou baixe o projeto**
   ```bash
   git clone <url-do-repositorio>
   cd quepasa
   ```

2. **Configure as variÃ¡veis de ambiente**
   ```bash
   cd docker
   cp .env.example .env
   # Edite o arquivo .env com suas configuraÃ§Ãµes especÃ­ficas
   ```

3. **Edite o arquivo `.env` com suas configuraÃ§Ãµes**
   - **DOMAIN**: Configure seu domÃ­nio (ex: `quepasa.seudominio.com`)
   - **EMAIL**: Defina seu email de administrador
   - **MASTERKEY**: Altere a chave mestra padrÃ£o por seguranÃ§a
   - **PASSWORD**: Defina uma senha forte
   - **DBDRIVER** / **DBDATABASE**: Definem o **banco de dados Ãºnico compartilhado** (tabelas quepasa_* da aplicaÃ§Ã£o + tabelas whatsmeow_* do store)
   - **DBPASSWORD**: Defina uma senha segura apenas quando usar PostgreSQL/MySQL (`DBDRIVER=postgres` ou `mysql`)
   - **SIGNING_SECRET**: Altere o segredo de assinatura padrÃ£o
   - **WEBSOCKETSSL**: Defina como `true` se usar HTTPS/SSL
   - **LOGLEVEL**: Ajuste o nÃ­vel de log (ERROR, WARN, INFO, DEBUG, TRACE)
   - **TZ**: Defina seu fuso horÃ¡rio

4. **Opcional: Revise o docker-compose.yml**
   - O arquivo compose agora usa variÃ¡veis de ambiente do `.env`
   - Inclui serviÃ§o PostgreSQL para o store do Whatsmeow
   - Mantenha as variÃ¡veis `DB*` alinhadas com o serviÃ§o do compose quando `DBDRIVER=postgres`
   - O banco interno do QuePasa continua local no cÃ³digo atual

5. **Construa e execute o container**
   ```bash
   # OpÃ§Ã£o 1: Construir e executar em um comando
   docker-compose up -d --build
   
   # OpÃ§Ã£o 2: Construir primeiro, depois executar
   docker-compose build
   docker-compose up -d
   
   # Usando sintaxe mais nova do Docker Compose
   docker compose up -d --build
   ```

6. **Verifique a instalaÃ§Ã£o**
   ```bash
   # Verificar status do container
   docker-compose ps
   
   # Verificar logs
   docker-compose logs -f quepasa
   ```

### Notas Importantes de ConfiguraÃ§Ã£o

- **Arquivo de Ambiente**: Todas as configuraÃ§Ãµes estÃ£o no arquivo `.env` para melhor gestÃ£o
- **Banco de Dados**: ServiÃ§o PostgreSQL incluÃ­do com configuraÃ§Ã£o automÃ¡tica
- **Primeira ConfiguraÃ§Ã£o**: Defina `ACCOUNTSETUP=true` para configuraÃ§Ã£o inicial
- **SeguranÃ§a**: Altere todas as senhas e segredos padrÃ£o no `.env`
- **SSL**: Defina `WEBSOCKETSSL=true` se usar HTTPS
- **Rede**: Usa rede interna do Docker `quepasa_network`
- **Portas**: Porta padrÃ£o 31000, configurÃ¡vel via `QUEPASA_EXTERNAL_PORT`

### VisÃ£o Geral das VariÃ¡veis de Ambiente

O arquivo `.env` contÃ©m todas as configuraÃ§Ãµes necessÃ¡rias organizadas em seÃ§Ãµes:
- **ConfiguraÃ§Ã£o BÃ¡sica**: DomÃ­nio, flags de setup, chave mestra
- **AutenticaÃ§Ã£o**: Email, senhas, configuraÃ§Ãµes de auth
- **Recursos WhatsApp**: Grupos, broadcasts, chamadas, recibos
- **Logging**: NÃ­veis de log para aplicaÃ§Ã£o e WhatsApp
- **Banco de Dados**: ConfiguraÃ§Ãµes de conexÃ£o PostgreSQL
- **Performance**: Cache, memÃ³ria, configuraÃ§Ãµes de sync
- **Debug**: VÃ¡rias opÃ§Ãµes de debugging

---

## Troubleshooting / SoluÃ§Ã£o de Problemas

### Common Issues / Problemas Comuns

#### Database Connection Issues / Problemas de ConexÃ£o com Banco
```bash
# Check database container
docker-compose ps

# View database logs
docker-compose logs db
```

#### Permission Issues / Problemas de PermissÃ£o
```bash
# Fix volume permissions
docker-compose down
sudo chown -R $USER:$USER quepasa_volume
docker-compose up -d
```

#### Port Conflicts / Conflitos de Porta
- Check if port 31000 is already in use
- Modify `QUEPASA_EXTERNAL_PORT` in .env file
- Update docker-compose.yml port mappings

#### Container Won't Start / Container NÃ£o Inicia
```bash
# Check detailed logs
docker-compose logs --details quepasa

# Rebuild without cache
docker-compose build --no-cache
docker-compose up -d
```

### Useful Commands / Comandos Ãšteis

```bash
# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v

# Restart specific service
docker-compose restart quepasa

# Access container shell
docker-compose exec quepasa sh

# View real-time logs
docker-compose logs -f --tail=100 quepasa
```

### Health Check / VerificaÃ§Ã£o de SaÃºde

The container includes a health check endpoint:
```
http://your-domain:31000/healthapi
```

This endpoint should return status information about the QuePasa service.
