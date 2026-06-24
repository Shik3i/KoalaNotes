# Development Guide

## Prerequisites

- **Node.js** 20+ and **npm**
- **Go** 1.23+
- **Docker** (optional, for container builds)
- **Git**

## Project Structure

```
/
  apps/web/          # Svelte/SvelteKit frontend
  server/            # Go backend
  docs/              # Documentation
  docker-compose.example.yml
  Dockerfile
  .github/workflows/ # CI/CD
```

## Quick Start

### Frontend

```bash
cd apps/web

# Install dependencies
npm install

# Start development server
npm run dev

# The app runs at http://localhost:5173

# Build for production
npm run build

# Preview production build
npm run preview
```

### Backend

```bash
cd server

# Run the server directly
go run ./cmd/server

# Or build and run
go build -o bin/server ./cmd/server
./bin/server

# The server runs at http://localhost:8080
# Health check: http://localhost:8080/healthz
# Version: http://localhost:8080/api/version
```

### Full Stack (Development)

Run both the frontend dev server and backend server concurrently:

**Terminal 1 — Backend:**
```bash
cd server && go run ./cmd/server
```

**Terminal 2 — Frontend:**
```bash
cd apps/web && npm run dev
```

The frontend dev server proxies API requests to the backend or you can
configure the API URL in the frontend's environment.

## Testing

### Frontend

```bash
cd apps/web
npm run check        # Type-check with svelte-check
npm run lint         # Lint with ESLint (if configured)
npm run test         # Run tests (if configured)
```

### Backend

```bash
cd server
go test ./...        # Run all tests
go vet ./...         # Static analysis
gofmt -l .           # Check formatting
```

## Linting and Formatting

### EditorConfig

The project uses [EditorConfig](https://editorconfig.org/). Most editors
have built-in or plugin support. See `.editorconfig` in the root.

### Frontend

- TypeScript via `svelte-check`
- Prettier or ESLint (to be configured)

### Backend

- `gofmt` for formatting
- `go vet` for static analysis
- `golangci-lint` (optional, for more comprehensive linting)

## Docker Development

```bash
# Build the Docker image
docker build -t koalanotes:dev .

# Run with Docker Compose
docker compose up -d

# Check logs
docker compose logs -f

# Stop
docker compose down
```

## Commit and PR Guidelines

See `CONTRIBUTING.md` for guidelines on commits, pull requests, and code
style.

## Environment Variables

| Variable         | Used By  | Default     | Description                  |
|------------------|----------|-------------|------------------------------|
| `KOALA_LISTEN_ADDR` | Server | `:8080`   | HTTP listen address          |
| `KOALA_DATA_DIR`    | Server | `/data`   | SQLite data directory        |

## Useful Commands

```bash
# Check git diff for whitespace issues
git diff --check

# Run all checks (from root)
cd apps/web && npm run check && cd ../..
cd server && go test ./... && go vet ./... && cd ..

# Build everything
cd apps/web && npm run build && cd ../..
cd server && go build -o bin/server ./cmd/server && cd ..
```
