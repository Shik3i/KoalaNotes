# Multi-stage build for KoalaNotes
# Stage 1: Build frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /src/apps/web
COPY apps/web/package*.json ./
RUN npm ci
COPY apps/web/ ./
RUN npm run build

# Stage 2: Build backend
FROM golang:1.24-alpine AS backend-builder
WORKDIR /src
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/koalanotes ./cmd/server

# Stage 3: Runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -h /app koalanotes

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /bin/koalanotes /app/koalanotes

# Copy frontend static files
COPY --from=frontend-builder /src/apps/web/build /app/static

# Create data directory for SQLite
RUN mkdir -p /data && chown koalanotes:koalanotes /data

USER koalanotes

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

ENV KOALA_LISTEN_ADDR=:8080
ENV KOALA_DATA_DIR=/data

# In the future, the server will serve the frontend static files.
# For now, the server just provides the API endpoints.
CMD ["./koalanotes"]
