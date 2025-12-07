# WhatsApp Support (genfity-wa-support)

Gateway service untuk WhatsApp API dengan subscription validation, rate limiting, dan request proxying.

## 📋 Deskripsi

Service ini berfungsi sebagai:
- **Gateway Proxy** - Meneruskan requests ke WhatsApp API (genfity-wa)
- **Subscription Validation** - Memvalidasi subscription user sebelum request
- **Rate Limiting** - Membatasi request per user/session
- **Webhook Handler** - Menerima dan memproses webhook events

## 🏗️ Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│                 │     │                 │     │                 │
│   Client/App    │────▶│   wa-support    │────▶│     wa-api      │
│                 │     │   (Gateway)     │     │   (WhatsApp)    │
│                 │     │                 │     │                 │
└─────────────────┘     └────────┬────────┘     └─────────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │                 │
                        │    PostgreSQL   │
                        │   (wa_support)  │
                        │                 │
                        └─────────────────┘
```

## 🚀 Quick Start

### Local Development
```bash
# Copy environment file
cp .env.example .env
# Edit .env with your values

# Start with Docker Compose
docker compose up -d --build

# Check logs
docker compose logs -f wa-support
```

### Production (Docker Swarm)
```bash
# Copy environment file
cp .env.example .env
# Edit .env with production values

# Deploy stack
docker stack deploy -c docker-compose.swarm.yml wa-support

# Check status
docker service ls
docker service logs wa-support_wa-support
```

## 📁 File Structure

```
wa-support/
├── .env.example              # Environment template
├── .env                      # Your environment (git ignored)
├── docker-compose.yml        # Local development
├── docker-compose.swarm.yml  # Production (uses GHCR images)
├── Dockerfile                # Build configuration
└── README.md                 # This file
```

## 🔧 Configuration

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | `secret` |
| `DB_NAME` | Primary database name | `wa_support` |
| `TRANSACTIONAL_DB_NAME` | Transaction database | `wa_support_tx` |
| `WA_SERVER_URL` | WhatsApp API URL | `http://wa-api:8080` |
| `WA_ADMIN_TOKEN` | Token untuk wa-api | `your_token` |
| `ADMIN_TOKEN` | Admin access token | `your_admin_token` |

### Optional Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GATEWAY_MODE` | `enabled` | Enable/disable gateway |
| `RATE_LIMIT_WINDOW` | `60` | Rate limit window (seconds) |
| `DEFAULT_RATE_LIMIT` | `100` | Max requests per window |
| `LOG_LEVEL` | `info` | Log verbosity |

## 🌐 Networks

Service ini terhubung ke:
- `wa-network` - Komunikasi dengan wa-api
- `infra-network` - Akses ke PostgreSQL

## 📡 Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /health` | Health check |
| `GET /` | Home page |
| `/wa/*` | Gateway proxy ke WhatsApp API |
| `/webhook/wa` | Webhook receiver |
| `GET /bulk/cron/process` | Cron job untuk bulk campaigns |

## 🔄 CI/CD

GitHub Actions workflow akan:
1. Build Docker image on push to main
2. Push to GitHub Container Registry (ghcr.io)
3. Tag dengan commit SHA dan `latest`

## 📊 Monitoring

Service exposed di Traefik:
- URL: `https://wa-support.govconnect.my.id`
- Metrics tersedia di `/metrics` (jika diaktifkan)
