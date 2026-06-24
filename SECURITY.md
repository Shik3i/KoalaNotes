# Security Policy

## Supported Versions

KoalaNotes is in early planning / foundation phase. No releases exist yet. Security
issues in the current `main` branch should be reported but are not considered
production incidents.

## Reporting a Vulnerability

**Do not open a public issue.**

Email the maintainers at the address listed in the repository or GitHub profile.

You should receive a response within 7 days. Please allow up to 72 hours for an
initial acknowledgment during weekends and holidays.

## Expectations

- This project is pre-production. We appreciate reports but cannot offer bounties.
- Responsible disclosure is expected: give maintainers time to address the issue
  before any public discussion.
- We will credit reporters in release notes unless anonymity is requested.

## Security Model

KoalaNotes is designed with a privacy-first, local-first architecture:

- **Local data** is stored in the browser (IndexedDB) and is not encrypted by
  default.
- **Server-bound data** is encrypted client-side before transmission. The server
  stores only encrypted blobs.
- **Plaintext content** never reaches the server under normal operation.
- **Accounts** are optional and only needed for sync/backup features.

See `docs/ENCRYPTION_AND_SYNC.md` for the full encryption model.

## Scope

The following are **in scope** for security reports:

- Data exposure on the server (plaintext leaks).
- Broken encryption implementation.
- Authentication/authorization bypass.
- Server-side injection vulnerabilities (SQL, command, etc.).
- Client-side XSS or CSRF in the web app.

The following are **out of scope**:

- Theoretical attacks on local-only data (user controls their machine).
- Issues in third-party reverse proxies or hosting environments.
- Social engineering.
- Denial of service (at this stage).
