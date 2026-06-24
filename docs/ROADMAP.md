# Roadmap

## Current Status: Foundation Phase

- [x] Project structure and documentation
- [x] Frontend skeleton (Svelte/SvelteKit)
- [x] Backend skeleton (Go)
- [x] Docker scaffolding
- [x] CI/CD workflows
- [ ] MVP Phase 1: Offline campaign notebook
- [ ] MVP Phase 2: Live session mode
- [ ] MVP Phase 3: Roles and visibility
- [ ] MVP Phase 4: Encrypted sync
- [ ] Post-MVP features

## Phase 1: Offline Campaign Notebook

**Target**: Q3-Q4 2026 (estimated)

- [ ] IndexedDB data layer (Dexie.js)
- [ ] Campaign CRUD
- [ ] Note/page CRUD with Markdown
- [ ] Wiki link (`[[page]]`) parsing and rendering
- [ ] Backlink computation
- [ ] Tag management and filtering
- [ ] TTRPG templates (NPC, Location, Quest, Item, Faction, Session)
- [ ] Search (full-text across campaign content)
- [ ] Human-readable export

## Phase 2: Live Session Mode

**Target**: TBD

- [ ] Session timer (start/stop)
- [ ] Live comment bar (always visible)
- [ ] Timeline entry storage with clock + session time
- [ ] Session timeline panel
- [ ] Session recap view
- [ ] Link entries to context (campaign, session, current note)

## Phase 3: Roles and Visibility

**Target**: TBD

- [ ] Campaign member roles (GM, Player, Observer)
- [ ] Note section visibility flags
- [ ] UI visibility indicators
- [ ] Role-based filtering in UI
- [ ] Data model ready for multi-user sync

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
