# Encryption and Sync

> **Status**: This document describes the **planned** encryption and sync
> architecture. Encryption is **not yet implemented** in the codebase. All
> current data is stored locally in plaintext.

## Design Philosophy

KoalaNotes follows a **zero-knowledge** model for server-synced data:

- The server **never** sees plaintext campaign content.
- Encryption and decryption happen **exclusively on the client**.
- The server stores only encrypted blobs and metadata required for sync.
- Account credentials are used only for authentication, not for content
  encryption (unless derived via key-stretching).

## Local Data (Current / Phase 1-3)

- **Storage**: IndexedDB
- **Encryption**: None (plaintext)
- **Rationale**: Local browser storage is considered secure within the user's
  device boundary. Adding local encryption adds complexity without meaningful
  security benefit for the MVP phases, since the threat model assumes the
  user controls their local machine.

Future versions **may** add optional local encryption, but it is not an MVP
requirement.

## Encrypted Sync Flow (Phase 4)

### 1. Account Setup

1. User creates an optional account (email + password).
2. Password is hashed client-side (or server-side via TLS) using bcrypt.
3. Server stores only the password hash. The plaintext password is **not**
   stored on the server.
4. A sync encryption key is derived from the account credentials or a
   separate sync passphrase (TBD: PBKDF2 or Argon2).

### 2. Data Encryption (Client-Side)

Before sending data to the server:

1. Campaign/note content and metadata are serialized to JSON.
2. A random symmetric key is generated per campaign (or per sync unit).
3. The symmetric key encrypts the content using AES-256-GCM (or
   XChaCha20-Poly1305 via libsodium).
4. The symmetric key is encrypted with the user's sync key (hybrid
   encryption, similar to envelope encryption).
5. The encrypted payload (ciphertext + encrypted key + IV/nonce) is
   sent to the server.

### 3. Server Storage

The server stores:

- Encrypted payload (opaque binary blob).
- Account ID (to associate blobs with users).
- Vector clock or version metadata for conflict resolution.
- Timestamps for sync coordination.

The server stores **nothing** in plaintext that could reveal campaign content:
- No plaintext note titles
- No plaintext NPC names
- No plaintext session notes
- No plaintext tags

### 4. Sync Pull

1. Client requests encrypted updates since last sync.
2. Server returns encrypted blobs with version metadata.
3. Client decrypts blobs using the sync key.
4. Client merges decrypted data into local IndexedDB (with conflict
   resolution).

### 5. Sharing (Future)

Sharing campaigns between users while maintaining E2EE:

- The sharing user encrypts the campaign's symmetric key with the
  recipient's public key (or a shared secret).
- The encrypted key is stored on the server and delivered to the
  recipient.
- The recipient decrypts the campaign key and can access the content.
- The server never sees the plaintext campaign key.

## Cryptographic Choices (Planned)

| Component           | Algorithm                    | Library              |
|---------------------|------------------------------|----------------------|
| Symmetric Encryption| AES-256-GCM / XChaCha20-Poly1305 | Web Crypto / libsodium |
| Key Derivation      | PBKDF2 / Argon2id            | Web Crypto / libsodium |
| Password Hashing    | bcrypt (server-side)         | Go stdlib / golang.org/x/crypto |
| Digital Signatures  | Ed25519 (for sharing)        | libsodium / Web Crypto |

## Threat Model

### What We Protect Against

- Server compromise: Attacker gains access to the server database. They see
  only encrypted blobs and account metadata (email, timestamps).
- Network eavesdropping: TLS protects data in transit.
- Server operator curiosity: The server operator cannot read campaign content.

### What We Do NOT Protect Against (by design)

- Local machine compromise: If an attacker has access to the user's unlocked
  machine, they can read IndexedDB data and potentially access the sync key
  if stored locally.
- User password loss: If the user loses their password or sync passphrase,
  encrypted server data is irrecoverable (by design — no backdoor).
- Metadata leakage: The server can see when sync occurs, how much data is
  synced, and account email addresses.

### What We Defer

- Local IndexedDB encryption (future consideration).
- Secure enclave / hardware-backed key storage (future consideration).
- Perfect forward secrecy for sync messages (future consideration).

## Security Notes

- **No encryption before Phase 4.** All encryption interfaces in the
  codebase are placeholders until Phase 4 implementation begins.
- **Mark placeholders clearly.** Any encryption-related code before Phase 4
  must include comments indicating it is not yet functional.
- **No fake security.** We will not implement "encryption" that does not
  actually protect data and present it as real.
- **External audit.** Before considering sync/encryption production-ready,
  the implementation should be audited by a qualified third party.
