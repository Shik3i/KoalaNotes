# KoalaNotes 🐨

**Privacy-first, local-first TTRPG campaign notebook.** Free and open source.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)
[![Status: Early Foundation](https://img.shields.io/badge/status-early%20foundation-orange)](#status)

KoalaNotes is a **campaign note-taking web app** optimized for tabletop RPGs.
It works offline, stores your data locally in your browser, and optionally
syncs with a self-hosted server using end-to-end encryption.

## What Makes It Different

- 🏠 **Local-first**: Full functionality without internet or account. Your
  data lives in your browser, not on someone else's server.
- 🔒 **Privacy-first**: When you choose to sync, your campaign content is
  encrypted client-side. The server never sees plaintext data.
- 🎯 **Session capture**: An always-visible **live comment bar** lets you
  capture moments during a session without breaking your flow.
- 🐳 **Self-hostable**: Single-container Docker deployment. No external
  services required.
- 📝 **Markdown wiki**: `[[Wiki Links]]`, backlinks, tags, and templates
  for NPCs, Locations, Quests, Items, and more.
- 🎲 **System-agnostic**: Works with any TTRPG system. No D&D/Hasbro-specific
  branding or mechanics.

## Status

🚧 **Early Foundation** — The project structure, documentation, and minimal
app/server skeletons are being established. There are **no stable releases**
yet. See [ROADMAP.md](./docs/ROADMAP.md) for planned phases.

### What Exists Today

- [x] Project structure and documentation
- [x] Frontend skeleton (Svelte/SvelteKit)
- [x] Backend skeleton (Go with `/healthz` and `/api/version`)
- [x] Docker deployment scaffolding
- [x] CI/CD workflows

### Coming in MVP Phase 1

- [ ] Offline campaign creation and management
- [ ] Markdown notes with wiki links and backlinks
- [ ] TTRPG templates (NPC, Location, Quest, etc.)
- [ ] Full-text search and export

## Quick Start (Development)

```bash
# Clone the repository
git clone git@github.com:Shik3i/KoalaNotes.git
cd KoalaNotes

# Run the frontend dev server
cd apps/web
npm install
npm run dev        # → http://localhost:5173

# In another terminal, run the backend
cd server
go run ./cmd/server  # → http://localhost:8080
```

See [DEVELOPMENT.md](./docs/DEVELOPMENT.md) for detailed instructions.

## Quick Start (Self-Hosting)

```bash
cp docker-compose.example.yml docker-compose.yml
docker compose up -d
# → http://localhost:8080
```

See [SELF_HOSTING.md](./docs/SELF_HOSTING.md) for production deployment.

## Documentation

| Document | Description |
|----------|-------------|
| [PRODUCT.md](./docs/PRODUCT.md) | Product vision, target users, non-goals |
| [ARCHITECTURE.md](./docs/ARCHITECTURE.md) | Frontend, backend, storage, sync overview |
| [MVP.md](./docs/MVP.md) | Phased MVP plan |
| [DATA_MODEL.md](./docs/DATA_MODEL.md) | Core entities and relationships |
| [ENCRYPTION_AND_SYNC.md](./docs/ENCRYPTION_AND_SYNC.md) | Planned encryption and sync model |
| [SELF_HOSTING.md](./docs/SELF_HOSTING.md) | Docker deployment guide |
| [DEVELOPMENT.md](./docs/DEVELOPMENT.md) | Local development setup |
| [AGENT_GUIDELINES.md](./docs/AGENT_GUIDELINES.md) | Rules for coding agents |
| [ROADMAP.md](./docs/ROADMAP.md) | Development phases and milestones |

## Core Principles

1. **Privacy First** — Local data stays local. Synced data is encrypted
   client-side.
2. **Local First** — Full functionality offline. No account required.
3. **Open Source** — MIT licensed. Free forever. Community owned.
4. **System Agnostic** — Works with any TTRPG system.
5. **Self-Hostable** — Run it yourself with Docker.
6. **Accessible** — Semantic HTML, keyboard navigation, responsive design.

## What KoalaNotes Is NOT

- ❌ A Virtual Tabletop (VTT)
- ❌ A dice roller
- ❌ An encounter builder
- ❌ A character sheet app
- ❌ An AI recap tool
- ❌ A cloud-required service

## License

[MIT](./LICENSE) — free to use, modify, and distribute.

## Contributing

KoalaNotes is in early stages. Contributions are welcome — see
[CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines. Please review the
[product boundaries](./docs/PRODUCT.md) before submitting PRs.
