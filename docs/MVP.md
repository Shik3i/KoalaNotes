# MVP Plan

This document outlines the Minimum Viable Product phases for KoalaNotes.
Each phase builds on the previous, delivering value incrementally.

## Phase 1: Offline Campaign Notebook

**Goal**: A working offline-only campaign note-taking app with TTRPG templates
and wiki linking.

### Features

- Create, rename, delete local campaigns (stored in IndexedDB).
- Create, edit, delete notes/pages within a campaign.
- Markdown editing and preview.
- TTRPG note templates:
  - NPC
  - Location
  - Quest
  - Item
  - Faction
  - Session
  - Session Recap
- `[[Wiki Links]]` between pages.
- Automatic backlinks.
- Tags on pages.
- Full-text search across campaigns and pages.
- Human-readable export (Markdown, structured JSON).

### Technical Scope

- Svelte/SvelteKit frontend (fully client-side, static export or SPA mode).
- IndexedDB via Dexie.js or similar wrapper.
- No backend required.
- No accounts or auth.

---

## Phase 2: Live Session Mode

**Goal**: Add the always-visible live comment bar and session timeline.

### Features (on top of Phase 1)

- Start/stop session timer (manual).
- Always-visible bottom quick-entry bar.
- Save live comments as **TimelineEntries**.
- Each entry stores:
  - Real clock timestamp.
  - Session elapsed time.
  - Associated campaign, session, and current note/page.
- Optional side/right timeline panel showing session entries.
- Session recap view aggregating timeline entries.

### Technical Scope

- Still offline/local-first.
- UI layout expands to accommodate timeline panel.
- Session state management (active session, timer).

---

## Phase 3: Roles and Visibility — ✅ Complete

**Goal**: Introduce role-based visibility for campaign content.

### Features (on top of Phase 2)

- Campaign member roles: GM, Player, Observer.
- GM sees everything.
- Note sections with visibility flags:
  - GM-only
  - Shared with players
  - Observer/read-only
  - Personal/private
- UI indicators for visibility levels.
- Prepare data model for future multi-user sharing.

### Technical Scope

- Local-only phase — no server auth yet.
- Visibility model implemented in data store and UI.
- Designed to extend to server-side enforcement later.

---

## Phase 4: Optional Account and Encrypted Sync

**Goal**: Enable optional accounts, encrypted server sync, and self-hosting.

### Features (on top of Phase 3)

- Optional account creation (email + password).
- Go backend with SQLite storage.
- Client-side encryption of campaign/note content before upload.
- Server stores only encrypted records/blobs.
- No plaintext campaign content on server.
- Manual sync/backup (triggered by user).
- Optional automatic sync (configurable interval or on-save).

### Technical Scope

- Go backend with REST API.
- SQLite for server-side encrypted blob storage.
- Client-side encryption using Web Crypto API or libsodium.
- Account auth (bcrypt for password hashing, JWT or session tokens).
- Docker self-hosting.

---

## Post-MVP / Future

Features planned for after the MVP phases:

- **Image/Attachment Support**: Strict quotas, external-hosting-friendly design.
  No built-in image hosting.
- **Encrypted Sharing**: Share specific campaigns or pages with other users
  while preserving E2EE.
- **Read-Only Links**: Public read-only links for campaign content (viewer
  does not need an account).
- **Multi-GM Campaigns**: Multiple users with GM-level access.
- **In-Game Time/Calendar**: Track in-game dates and calendars alongside
  real-world timelines.
- **Layout Presets**: User-customizable panel layouts.
- **Plugin/Import Integrations**: Import from other tools, plugin system.
- **Desktop Wrapper (Tauri)**: Optional native desktop app via Tauri, though
  the web app should remain fully functional.

## MVP Non-Goals

These are explicitly deferred or excluded:

- VTT features (maps, tokens, fog of war)
- Dice rolling
- Encounter builders
- Character sheets
- AI features
- Real-time collaborative editing
- Rich-text/block editor (Markdown is first-class)
- Cloud-required features
- D&D/Hasbro-specific content or branding
