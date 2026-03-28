# vh-srv-events

Events management microservice for the Virtual Home platform. Handles event registration, participation tracking, broadcast scheduling, and participant notifications.

## Tech Stack

- **Go 1.21** with [Gin](https://github.com/gin-gonic/gin) HTTP framework
- **PostgreSQL** (9.6 staging / 12 production)
- **Keycloak** OIDC authentication
- **Sentry** error tracking
- **SendGrid** email notifications

## API

All endpoints are under `/v1` and require Keycloak authentication.

| Resource | Description |
|----------|-------------|
| `/participant` | Participant profiles (create, update, lookup by email/keycloak ID) |
| `/event` | Event CRUD with registration status, dates, audience |
| `/participation-status` | Track who is attending which event |
| `/participation-option` | Attendance options (attending, not attending, etc.) |
| `/platform` | Broadcast platforms (YouTube, Teams, etc.) |
| `/audience` | Event audience segments |
| `/item` | Event sessions/items |
| `/broadcasturl` | Stream URLs for items |
| `/notification/event` | Send event email notifications |
| `/analytics/participants` | Participation analytics |
| `/health` | Health check (includes DB connectivity) |

## Development

```bash
# Start local PostgreSQL
task dev:standalone

# Run the service
task dev

# Run tests
task test
```

See `Taskfile.yml` for all available commands.

## Configuration

The service is configured via environment variables (loaded from `.env`):

| Variable | Description |
|----------|-------------|
| `APP_PORT` | Server port (default: 8080) |
| `APP_MODE` | Gin mode: debug/release |
| `APP_ENV` | Environment: dev/staging/production |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_DATABASE` | PostgreSQL connection |
| `KEYCLOAK_SERVER_URL`, `KEYCLOAK_REALM`, `KEYCLOAK_CLIENT_ID`, `KEYCLOAK_CLIENT_SECRET` | OIDC auth |
| `SENTRY_DSN` | Error tracking |

## Deployment

CI/CD runs via GitHub Actions (`workflow_dispatch`):

1. **Test** - Go tests with PostgreSQL 12 service
2. **Build** - Docker image pushed to `ghcr.io/bnei-baruch/vh-srv-events`
3. **Deploy** - Image pulled and started on target VM via SSH

Trigger manually from the Actions tab, selecting `staging` or `production`.
