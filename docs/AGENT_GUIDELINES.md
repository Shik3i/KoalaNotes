# Agent Guidelines

This document contains rules and conventions for coding agents (AI or human)
working on the KoalaNotes codebase. Follow these strictly.

## Product Boundaries

### DO

- Implement offline-first, local-first features.
- Use Markdown as the primary note format.
- Use system-agnostic TTRPG language: NPC, Location, Quest, Session, Faction,
  Item, Campaign, GM, Player, Observer.
- Design for the core layout: left sidebar, center note area, right timeline
  panel (optional), bottom live comment bar.
- Keep the app fast and lightweight.
- Use semantic HTML, keyboard navigation, visible focus states, and labels for
  form controls.
- Document planned-but-not-implemented features clearly.
- Mark encryption-related code as placeholder until Phase 4.
- Write boring, maintainable code over clever abstractions.
- Add comments where future decisions matter.

### DO NOT

- Do not add VTT features (maps, tokens, fog of war).
- Do not add dice rolling.
- Do not add encounter builders or combat trackers.
- Do not add character sheet management.
- Do not add AI-powered features (recap, generation, etc.).
- Do not use DnD/Hasbro-specific naming or branding.
- Do not add analytics, telemetry, tracking, ads, or external CDNs.
- Do not add login/auth unless required for the current phase.
- Do not implement fake/insecure encryption and present it as real.
- Do not store real secrets in the repository.
- Do not add image/attachment upload handling until post-MVP phases.
- Do not claim production readiness without explicit approval.

## Code Conventions

### General

- Follow the `.editorconfig` settings.
- TypeScript for frontend code.
- Go idioms for backend code.
- Prefer explicit, readable code over terse one-liners.
- Use meaningful variable and function names.
- Keep functions small and focused.
- Add JSDoc comments for public APIs in TypeScript.
- Add doc comments for exported functions in Go.

### Frontend (Svelte/SvelteKit)

- Use TypeScript throughout.
- Prefer Svelte 5 runes syntax (`$state`, `$derived`, `$effect`) if using
  Svelte 5.
- Use SvelteKit's file-based routing.
- Keep components small and composable.
- Use Svelte stores for shared state.
- Accessible by default: `aria-*` attributes, `role` attributes, semantic
  elements, keyboard handlers.
- Responsive design: support mobile, tablet, and desktop.
- No heavy CSS frameworks unless explicitly approved (plain CSS or
  Tailwind/Open-Props are acceptable later).

### Backend (Go)

- Standard library first; add dependencies only when justified.
- Use `net/http` with a lightweight router if needed (chi, mux).
- Separate handlers, middleware, and storage into `internal/` packages.
- SQLite via `modernc.org/sqlite` (pure Go, no CGo dependency).
- Handle errors explicitly; never ignore errors.
- Use structured logging (stdlib `log/slog`).
- Write table-driven tests.
- Graceful shutdown on SIGTERM/SIGINT.

## Documentation

- Keep docs in sync with code.
- The `docs/` directory is the source of truth for architecture, data model,
  and product decisions.
- When adding a new feature, update relevant docs.
- Use Markdown for all documentation.

## Git Practices

- One logical change per commit.
- Write concise, present-tense commit messages.
- Reference issues where applicable.
- Never commit secrets, `.env` files, or generated binaries.

## Safety Rules

- Never introduce external telemetry or tracking.
- Never introduce a hard dependency on a hosted service.
- The app must work offline without an account.
- Encryption must be real and verifiable; never fake.
- When in doubt about a feature scope, refer to `docs/PRODUCT.md` and
  `docs/MVP.md`.

## Phase Awareness

Be aware of which MVP phase the project is in:

- **Phase 1 (current)**: Foundation + offline campaign notebook. No backend
  features beyond health/version endpoints.
- **Phase 2**: Live session mode (planned).
- **Phase 3**: Roles and visibility (planned).
- **Phase 4**: Encryption and sync (planned).

Do not implement features from future phases unless explicitly instructed.
