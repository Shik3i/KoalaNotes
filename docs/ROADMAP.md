# Roadmap

## Current Status: Phase 3 Complete

- [x] Project structure and documentation
- [x] Frontend skeleton (Svelte/SvelteKit)
- [x] Backend skeleton (Go)
- [x] Docker scaffolding
- [x] CI/CD workflows
- [x] MVP Phase 1: Offline campaign notebook
- [x] MVP Phase 2: Live session mode
- [x] MVP Phase 3: Roles and visibility
- [ ] MVP Phase 4: Encrypted sync
- [ ] Post-MVP features

## Phase 1: Offline Campaign Notebook — ✅ Complete

- [x] IndexedDB data layer (Dexie.js)
- [x] Campaign CRUD
- [x] Note/page CRUD with Markdown
- [x] Wiki link (`[[page]]`) parsing and rendering
- [x] Backlink computation
- [x] Tag management and filtering
- [x] TTRPG templates (NPC, Location, Quest, Item, Faction, Session)
- [x] Search (full-text across campaign content)
- [x] Human-readable export

## Phase 2: Live Session Mode — ✅ Complete

- [x] Session timer (start/stop)
- [x] Live comment bar (always visible)
- [x] Timeline entry storage with clock + session time
- [x] Session timeline panel
- [x] Session recap view (compile timeline entries into a recap note)
- [x] Link entries to context (show which note a comment was made on)

## Phase 3: Roles and Visibility — ✅ Complete

- [x] Campaign member roles (GM, Player, Observer)
- [x] Note section visibility flags
- [x] UI visibility indicators
- [x] Role-based filtering in UI
- [x] Data model ready for multi-user sync

## Phase 4: Encrypted Sync

**Target**: TBD

- [ ] Account system (email + password)
- [ ] Go backend API expansion
- [ ] SQLite server storage
- [ ] Client-side encryption (Web Crypto / libsodium)
- [ ] Encrypted push/pull sync
- [ ] Manual sync trigger
- [ ] Automatic sync option
- [ ] Server stores only encrypted blobs

## Post-MVP

**Target**: TBD, community-driven

- [ ] Image/attachment support (quotas, external hosting friendly)
- [ ] Encrypted campaign sharing
- [ ] Read-only public links
- [ ] Multi-GM campaigns
- [ ] In-game time/calendar
- [ ] Layout presets
- [ ] Plugin/import system
- [ ] Desktop wrapper (Tauri)

---

## Release History

No releases yet. See [CHANGELOG.md](../CHANGELOG.md) for development progress.
