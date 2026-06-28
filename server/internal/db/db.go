package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// BlobRecord represents an encrypted payload stored on the server.
type BlobRecord struct {
	ID              string `json:"id"`
	AccountID       string `json:"account_id"`
	CampaignKeyID   string `json:"campaign_key_id"`
	EncryptedPayload string `json:"encrypted_payload"`
	VectorClock     string `json:"vector_clock,omitempty"`
	CreatedAt       string `json:"created_at"`
}

// DB wraps the SQLite connection and provides data access methods.
type DB struct {
	conn *sql.DB
}

// Open opens (or creates) the SQLite database at the given path and runs migrations.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	conn.SetMaxOpenConns(1) // SQLite doesn't support concurrent writes

	if err := migrate(conn); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close shuts down the database connection.
func (d *DB) Close() error {
	return d.conn.Close()
}

func migrate(conn *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS blob_records (
			id TEXT PRIMARY KEY,
			account_id TEXT NOT NULL,
			campaign_key_id TEXT NOT NULL,
			encrypted_payload TEXT NOT NULL,
			vector_clock TEXT,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			FOREIGN KEY (account_id) REFERENCES accounts(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_blob_account ON blob_records(account_id)`,
	}

	for _, q := range queries {
		if _, err := conn.Exec(q); err != nil {
			return fmt.Errorf("exec %q: %w", q[:40], err)
		}
	}
	return nil
}

// ---- Accounts ----

// CreateAccount inserts a new account. id should be a UUID.
func (d *DB) CreateAccount(id, email, passwordHash string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := d.conn.Exec(
		`INSERT INTO accounts (id, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		id, email, passwordHash, now, now,
	)
	return err
}

// GetAccountByEmail returns the account's id and password hash for the given email.
func (d *DB) GetAccountByEmail(email string) (id, passwordHash string, err error) {
	err = d.conn.QueryRow(
		`SELECT id, password_hash FROM accounts WHERE email = ?`, email,
	).Scan(&id, &passwordHash)
	return
}

// AccountExists checks if an account with the given email exists.
func (d *DB) AccountExists(email string) (bool, error) {
	var count int
	err := d.conn.QueryRow(`SELECT COUNT(*) FROM accounts WHERE email = ?`, email).Scan(&count)
	return count > 0, err
}

// ---- Blob Records ----

// UpsertBlob inserts or replaces a blob record.
func (d *DB) UpsertBlob(blob BlobRecord) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := d.conn.Exec(
		`INSERT INTO blob_records (id, account_id, campaign_key_id, encrypted_payload, vector_clock, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
			encrypted_payload = excluded.encrypted_payload,
			vector_clock = excluded.vector_clock,
			created_at = excluded.created_at`,
		blob.ID, blob.AccountID, blob.CampaignKeyID, blob.EncryptedPayload, blob.VectorClock, now,
	)
	return err
}

// UpsertBlobs inserts or replaces multiple blobs in a transaction.
func (d *DB) UpsertBlobs(blobs []BlobRecord) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT INTO blob_records (id, account_id, campaign_key_id, encrypted_payload, vector_clock, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
			encrypted_payload = excluded.encrypted_payload,
			vector_clock = excluded.vector_clock,
			created_at = excluded.created_at`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	for _, b := range blobs {
		if _, err := stmt.Exec(b.ID, b.AccountID, b.CampaignKeyID, b.EncryptedPayload, b.VectorClock, now); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetBlobs returns all blob records for the given account, optionally filtered by a minimum created_at.
func (d *DB) GetBlobs(accountID, since string) ([]BlobRecord, error) {
	var rows *sql.Rows
	var err error

	if since != "" {
		rows, err = d.conn.Query(
			`SELECT id, account_id, campaign_key_id, encrypted_payload, COALESCE(vector_clock, ''), created_at
			 FROM blob_records
			 WHERE account_id = ? AND created_at > ?
			 ORDER BY created_at ASC`,
			accountID, since,
		)
	} else {
		rows, err = d.conn.Query(
			`SELECT id, account_id, campaign_key_id, encrypted_payload, COALESCE(vector_clock, ''), created_at
			 FROM blob_records
			 WHERE account_id = ?
			 ORDER BY created_at ASC`,
			accountID,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blobs []BlobRecord
	for rows.Next() {
		var b BlobRecord
		if err := rows.Scan(&b.ID, &b.AccountID, &b.CampaignKeyID, &b.EncryptedPayload, &b.VectorClock, &b.CreatedAt); err != nil {
			return nil, err
		}
		blobs = append(blobs, b)
	}
	return blobs, rows.Err()
}

// GetBlobIDs returns the set of blob IDs the server has for an account (for conflict detection).
func (d *DB) GetBlobIDs(accountID string) (map[string]bool, error) {
	rows, err := d.conn.Query(`SELECT id FROM blob_records WHERE account_id = ?`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids[id] = true
	}
	return ids, rows.Err()
}
