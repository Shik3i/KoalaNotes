# Self-Hosting

KoalaNotes is designed to be fully self-hostable. You can run the entire
stack on your own infrastructure with no dependency on external services.

## Architecture for Self-Hosting

The self-hosted deployment consists of:

- **KoalaNotes Server** (Go binary): Serves the API and embedded frontend
  static files.
- **SQLite Database**: Stored as a file on a persistent Docker volume.
- **Reverse Proxy (optional but recommended)**: nginx, Caddy, or Traefik
  for TLS termination.

```
[Browser] ---TLS---> [Reverse Proxy] ---HTTP---> [KoalaNotes :8080] ---> [SQLite /data/koalanotes.db]
```

## Docker Compose (Recommended)

The repository includes a `docker-compose.example.yml` for easy setup.

### Quick Start

```bash
# Copy and customize the example
cp docker-compose.example.yml docker-compose.yml

# Edit docker-compose.yml to set your configuration
# (domain, secrets, etc.)

# Start the service
docker compose up -d
```

### Environment Variables

| Variable              | Required | Default                | Description                              |
|-----------------------|----------|------------------------|------------------------------------------|
| `KOALA_LISTEN_ADDR`   | No       | `:8080`                | Server listen address                    |
| `KOALA_DATA_DIR`      | No       | `/data`                | Directory for SQLite database            |
| `KOALA_JWT_SECRET`    | No       | (generated if empty)   | Secret for JWT signing (future)          |
| `KOALA_CORS_ORIGINS`  | No       | `*`                    | Allowed CORS origins (future)            |

### Volumes

Persist the `/data` directory to retain your SQLite database across container
restarts:

```yaml
volumes:
  koalanotes-data:
```

## Manual Deployment (Without Docker)

1. **Build the frontend**:
   ```bash
   cd apps/web
   npm install
   npm run build
   ```

2. **Build the backend**:
   ```bash
   cd server
   go build -o bin/server ./cmd/server
   ```

3. **Run**:
   ```bash
   KOALA_DATA_DIR=./data ./server/bin/server
   ```

   Or serve the frontend separately with a static file server and configure
   the API URL in the frontend.

## Production Considerations

### TLS

Always use TLS in production. The recommended setup:

- Use Caddy or nginx as a reverse proxy with Let's Encrypt.
- The Go server itself does not handle TLS directly.

Example Caddy configuration:

```
koalanotes.example.com {
    reverse_proxy localhost:8080
}
```

### Database Backups

The SQLite database is a single file. Back up by copying the file:

```bash
cp /data/koalanotes.db /backups/koalanotes-$(date +%Y%m%d).db
```

Ensure the database is not being written during backup (stop the service or
use SQLite's `.backup` command).

### Resource Limits

KoalaNotes is lightweight:

- **Memory**: ~50-100 MB for the Go server under normal load.
- **CPU**: Minimal. Single-core is sufficient for personal/hobby use.
- **Storage**: Depends on usage. SQLite database grows with encrypted blobs.
  Plan for at least 1 GB for personal use, more for heavily shared instances.

## No External Dependencies

KoalaNotes' normal operation does not require:

- Cloud databases (PostgreSQL, MySQL as-a-service)
- Object storage (S3, etc.)
- External authentication providers (OAuth, etc.)
- Monitoring/telemetry services
- CDN services

Everything runs within the single container (or a small set of containers
with a reverse proxy).
