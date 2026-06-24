# Architecture

## Overview

KoalaNotes follows a **local-first** architecture with an **optional encrypted
sync layer**.

```
+-----------------------+       +---------------------------+
|                       |       |                           |
|  Browser (SvelteKit)  |<----->|  Go Backend (optional)    |
|                       | sync  |                           |
|  +------------------+ |       |  +---------------------+  |
|  |   IndexedDB      | |       |  |   SQLite            |  |
|  |   (plaintext)    | |       |  |   (encrypted blobs) |  |
|  +------------------+ |       |  +---------------------+  |
|                       |       |                           |
|  +------------------+ |       |  +---------------------+  |
|  |   Encryption     | |       |  |   Auth / Accounts   |  |
|  |   (Web Crypto)   | |       |  |   (optional)        |  |
|  +------------------+ |       |  +---------------------+  |
|                       |       |                           |
+-----------------------+       +---------------------------+
```

## Frontend (`apps/web`)

### Technology

- **Framework**: SvelteKit (in SPA mode or static adapter for offline-first)
- **Language**: TypeScript
- **Styling**: CSS (no heavy framework; consider Tailwind or Open-Props later)
- **Local Storage**: IndexedDB (via Dexie.js)
- **Encryption**: Web Crypto API (SubtleCrypto) or libsodium.js

### Component Architecture (Planned)

```
src/
  lib/
    components/
      layout/
        AppShell.svelte        # Main layout: sidebar + main + timeline + comment bar
        Sidebar.svelte          # Campaign/page navigation tree
        TimelinePanel.svelte    # Session timeline (Phase 2+)
        LiveCommentBar.svelte   # Bottom quick-entry bar
      campaign/
        CampaignList.svelte
        CampaignCreate.svelte
      note/
        NoteEditor.svelte       # Markdown editor + preview
        NoteList.svelte
        WikiLink.svelte         # [[Wiki Link]] handling
      templates/
        TemplatePicker.svelte
        NpcTemplate.svelte
        LocationTemplate.svelte
        ...
      session/
        SessionTimer.svelte
        TimelineEntry.svelte
      common/
        MarkdownPreview.svelte
        TagInput.svelte
    stores/
      campaignStore.ts
      noteStore.ts
      sessionStore.ts
      timelineStore.ts
      uiStore.ts
    db/
      database.ts              # Dexie.js schema and initialization
      migrations.ts            # IndexedDB upgrade helpers
    crypto/
      encryption.ts            # Web Crypto API wrappers (placeholder)
      keys.ts                  # Key generation/storage
    types/
      models.ts                # TypeScript type definitions
    utils/
      slug.ts
      markdown.ts
      export.ts
  routes/
    +layout.svelte
    +page.svelte               # Landing page / campaign list
    campaign/
      [campaignId]/
        +page.svelte           # Campaign overview
        notes/
          [noteId]/
            +page.svelte       # Note editor view
```

### Data Flow (Local-First)

1. User interacts with Svelte components.
2. Components dispatch actions to Svelte stores.
3. Stores interact with Dexie.js (IndexedDB wrapper).
4. Data is persisted locally as plaintext JSON.
5. UI updates reactively via Svelte's store subscriptions.

When sync is implemented (Phase 4):
1. A sync service periodically or manually pushes encrypted snapshots to the
   Go backend.
2. Encrypted data is received from the server and decrypted client-side before
   merging into IndexedDB.

## Backend (`server/`)

### Technology

- **Language**: Go
- **Framework**: Standard library `net/http` (with `gorilla/mux` or `chi` if
  routing complexity grows)
- **Database**: SQLite (via `modernc.org/sqlite` or `mattn/go-sqlite3`)
- **Auth**: bcrypt for passwords, JWT for sessions (future)

### Package Structure (Planned)

```
server/
  cmd/
    server/
      main.go                 # Entry point
  internal/
    handler/
      health.go               # /healthz
      version.go              # /api/version
      auth.go                 # Future: /api/auth/*
      sync.go                 # Future: /api/sync/*
      share.go                # Future: /api/share/*
    middleware/
      logging.go
      cors.go
      auth.go                 # Future: JWT validation
    store/
      sqlite.go               # SQLite connection and queries
      migrations.go           # Schema migrations
      accounts.go             # Account CRUD
      blobs.go                # Encrypted blob CRUD
  migrations/
    001_initial.sql           # Future: initial schema
  go.mod
  go.sum
```

### API Endpoints

#### Current (Foundation)

| Method | Path          | Description                 | Status   |
|--------|---------------|-----------------------------|----------|
| GET    | `/healthz`    | Health check                | Done     |
| GET    | `/api/version`| Version information         | Done     |

#### Planned (Future Phases)

| Method | Path                   | Description              | Phase |
|--------|------------------------|--------------------------|-------|
| POST   | `/api/auth/register`   | Create account           | 4     |
| POST   | `/api/auth/login`      | Login                    | 4     |
| POST   | `/api/auth/refresh`    | Refresh JWT              | 4     |
| GET    | `/api/sync/pull`       | Pull encrypted updates   | 4     |
| POST   | `/api/sync/push`       | Push encrypted updates   | 4     |
| POST   | `/api/share`           | Create share link        | Post  |
| GET    | `/api/share/:id`       | Access shared content    | Post  |

### Data Storage

The server stores **only** encrypted blobs. It never processes plaintext
campaign content. See `ENCRYPTION_AND_SYNC.md` for details.

```sql
-- Future schema sketch (not implemented)
CREATE TABLE accounts (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE blob_records (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    campaign_key_id TEXT NOT NULL,
    encrypted_payload BLOB NOT NULL,
    vector_clock TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (account_id) REFERENCES accounts(id)
);
```

## Security Model

1. **Local Data**: Plaintext in IndexedDB. Secure as the user's browser.
2. **Encryption at Rest (Server)**: All campaign content is encrypted
   client-side before upload. The server stores only ciphertext.
3. **Encryption in Transit**: TLS enforced (handled by reverse proxy in
   self-hosted deployments).
4. **Authentication**: Optional. Only needed for sync. Password hashed with
   bcrypt. JWTs for session management.
5. **Key Management**: Encryption keys derived from account password (or a
   separate sync passphrase). Keys never leave the client.

## Deployment

See `SELF_HOSTING.md` for Docker deployment. See `Dockerfile` and
`docker-compose.example.yml` in the repository root.

- The frontend is built as static files and served by the Go backend (or a
  reverse proxy like nginx/Caddy).
- The Go backend connects to a local SQLite file in a mounted volume.
- No external services required.
