# Product Vision

KoalaNotes is a free, open-source, privacy-first, local-first, TTRPG-optimized
campaign note-taking web app.

## What KoalaNotes Is

KoalaNotes is a **campaign notebook**. It helps Game Masters and players organize
their tabletop RPG notes, capture live session moments, and link everything
together in a Markdown wiki.

Think of it as closer to *The Goblin's Notebook* than to *Obsidian*, but with a
strong local-first philosophy, optional encrypted sync, and a distinctive
always-visible live comment bar optimized for TTRPG session capture.

### Core Capabilities

- **Campaign Wiki**: Create and link notes using `[[Wiki Links]]`, Markdown, tags,
  and backlinks.
- **Live Session Capture**: An always-accessible quick-entry bar at the bottom of
  the app for capturing moments during a session without breaking flow.
- **Session Timeline**: Every live entry is stored with real clock time, session
  timer elapsed time, and context (campaign, session, current note).
- **TTRPG Templates**: Pre-built note templates for common entities: NPC,
  Location, Quest, Item, Faction, Session, Session Recap.
- **Roles & Visibility**: GM-only content, player-shared content, observer
  read-only areas, and personal/private sections within a shared campaign.
- **Local-First**: Full offline use. All data stored locally in the browser
  (IndexedDB). No account required.
- **Optional Sync**: Encrypted, privacy-preserving sync/backup via a self-hosted
  server. The server never sees plaintext campaign content.
- **Self-Hostable**: Run the sync server yourself. No dependency on a hosted
  service.
- **Human-Readable Export**: Export your data in a format you can read and use
  outside KoalaNotes.

## What KoalaNotes Is NOT

KoalaNotes intentionally does **not** include:

- Virtual Tabletop (VTT) features (maps, tokens, fog of war)
- Dice rolling
- Encounter builders or combat trackers
- Character sheet management
- AI-powered recap or content generation
- Cloud-required note hosting
- File/image backup dumping ground

## Target Users

- **Game Masters** who want organized campaign notes and live session capture
  without complex tools.
- **Players** who want personal notes linked to campaign context.
- **Groups** who want shared campaign wikis with role-based visibility.
- **Privacy-conscious users** who want local-first tools that work offline.
- **Self-hosters** who want to own their data and infrastructure.

## Core Principles

1. **Privacy First**: Local data stays local. Server-synced data is encrypted
   client-side. The server never sees plaintext content.
2. **Local First**: Full functionality without internet or account. Sync is
   an enhancement, not a requirement.
3. **Open Source**: MIT licensed. Free forever. Community owned.
4. **TTRPG Optimized**: Designed for tabletop campaign workflows, not generic
   note-taking.
5. **System Agnostic**: Works with any TTRPG system. No D&D-specific branding
   or mechanics.
6. **Fast & Lightweight**: The app should feel responsive. No bloated editors
   or complex UI.
7. **Accessible**: Semantic HTML, keyboard navigation, visible focus states,
   responsive design.
8. **Self-Hostable**: Simple Docker deployment. No external SaaS dependencies.

## User Experience Vision

The app layout communicates shape and purpose from the start:

```
+------------------+---------------------------+-------------------+
|                  |                           |                   |
|  Campaign Tree   |    Note Editor /          |   Timeline /      |
|  Page List       |    Markdown Preview       |   Context Panel   |
|                  |                           |                   |
|                  |                           |                   |
+------------------+---------------------------+-------------------+
|                                                                   |
|  Live Comment Bar: [Type a quick note...]      [00:42:15] [Save] |
|                                                                   |
+-------------------------------------------------------------------+
```

- **Left**: Campaign and page navigation sidebar.
- **Center**: Main note editor with Markdown support.
- **Right (optional)**: Timeline view, backlinks, or context panel.
- **Bottom**: Always-visible live comment bar with session timer.

The live comment bar is a **core differentiator**, not an afterthought. It must
always be accessible and fast, even during heavy editing.
