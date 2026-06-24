# Contributing

Thank you for your interest in KoalaNotes!

## Project Status

KoalaNotes is in **early planning / foundation phase**. There are no stable
releases yet. The current focus is on establishing the project structure,
documentation, and core offline/local-first note-taking and session capture.

## How to Contribute

### 1. Open an Issue First

- For bugs, feature requests, or significant discussions, open an issue before
  submitting a pull request.
- This helps avoid duplicate work and ensures alignment with the project
  direction.

### 2. Keep Pull Requests Focused

- One logical change per PR.
- Link the relevant issue(s).
- Update documentation if your change affects user-facing behavior or
  architecture.

### 3. Code Style

- Follow the [EditorConfig](./.editorconfig) settings.
- **Frontend (Svelte/SvelteKit):** Use TypeScript. Follow SvelteKit best
  practices.
- **Backend (Go):** Follow standard Go idioms. Use `gofmt` and `go vet`.
- Prefer boring, maintainable code over clever abstractions.

### 4. Commit Messages

- Use present tense ("Add feature" not "Added feature").
- Keep the first line under 72 characters.
- Reference issue numbers where applicable.

### 5. Testing

- Add tests for new functionality where practical.
- Ensure existing tests pass before submitting.

### 6. Sign-off

All commits must include a `Signed-off-by` line:

```
Signed-off-by: Your Name <your@email.example>
```

This certifies that you have the right to submit the work under the project's
license (MIT).

## A Note on Dependencies

KoalaNotes has a strict no-telemetry policy. The project does not include any
runtime analytics, tracking, or data collection. Some build tools (e.g., Vite
via `@sveltejs/vite-plugin-svelte`) pull in `@opentelemetry/api` as an optional
transitive dependency for build-time instrumentation. This is **not** used at
runtime and does not send data anywhere — it is a standard part of the Vite
ecosystem. If you have questions about a specific dependency, open an issue.

## Development Setup

See [DEVELOPMENT.md](./docs/DEVELOPMENT.md) for local development instructions.

## Code of Conduct

All participants must follow the [Code of Conduct](./CODE_OF_CONDUCT.md).

## Product Boundaries

Before contributing, please review [PRODUCT.md](./docs/PRODUCT.md) to understand
what KoalaNotes is and is not. In particular:

- KoalaNotes is **not** a VTT, dice roller, encounter builder, or character
  sheet app.
- Use system-agnostic TTRPG language (NPC, Location, Quest, Session, etc.).
- Do not add DnD/Hasbro-specific naming or branding.
- Do not add analytics, telemetry, tracking, or external CDNs.
