package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "modernc.org/sqlite"
)

// BlobRecord represents an encrypted payload stored on the server.
type BlobRecord struct {
	ID               string `json:"id"`
	AccountID        string `json:"account_id"`
	CampaignKeyID    string `json:"campaign_key_id"`
	EncryptedPayload string `json:"encrypted_payload"`
	VectorClock      string `json:"vector_clock,omitempty"`
	CreatedAt        string `json:"created_at"`
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

// Ping checks if the database connection is alive.
func (d *DB) Ping(ctx context.Context) error {
	return d.conn.PingContext(ctx)
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
		`CREATE INDEX IF NOT EXISTS idx_blob_account_created ON blob_records(account_id, created_at)`,
	}

	for _, q := range queries {
		if _, err := conn.Exec(q); err != nil {
			return fmt.Errorf("exec %q: %w", truncate(q, 60), err)
		}
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// ---- Accounts ----

// CreateAccount inserts a new account. id should be a UUID.
func (d *DB) CreateAccount(ctx context.Context, id, email, passwordHash string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := d.conn.ExecContext(ctx,
		`INSERT INTO accounts (id, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		id, email, passwordHash, now, now,
	)
	return err
}

// GetAccountByEmail returns the account's id and password hash for the given email.
func (d *DB) GetAccountByEmail(ctx context.Context, email string) (id, passwordHash string, err error) {
	err = d.conn.QueryRowContext(ctx,
		`SELECT id, password_hash FROM accounts WHERE email = ?`, email,
	).Scan(&id, &passwordHash)
	return
}

// ---- Blob Records ----

// UpsertBlobs inserts or replaces multiple blobs in a transaction.
func (d *DB) UpsertBlobs(ctx context.Context, blobs []BlobRecord) error {
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	commit := false
	defer func() {
		if !commit {
			if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
				slog.Error("rollback failed", "error", err)
			}
		}
	}()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO blob_records (id, account_id, campaign_key_id, encrypted_payload, vector_clock, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
			encrypted_payload = excluded.encrypted_payload,
			vector_clock = excluded.vector_clock`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, b := range blobs {
		now := time.Now().UTC().Format(time.RFC3339)
		if _, err := stmt.ExecContext(ctx, b.ID, b.AccountID, b.CampaignKeyID, b.EncryptedPayload, b.VectorClock, now); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	commit = true
	return nil
}

// validateSince returns an error if since is not a valid RFC3339 string.
func validateSince(since string) error {
	if since == "" {
		return nil
	}
	_, err := time.Parse(time.RFC3339, since)
	return err
}

// GetBlobs returns all blob records for the given account, optionally filtered by a minimum created_at.
func (d *DB) GetBlobs(ctx context.Context, accountID, since string) ([]BlobRecord, error) {
	if err := validateSince(since); err != nil {
		return nil, fmt.Errorf("invalid since parameter: %w", err)
	}

	var rows *sql.Rows
	var err error

	if since != "" {
		rows, err = d.conn.QueryContext(ctx,
			`SELECT id, account_id, campaign_key_id, encrypted_payload, COALESCE(vector_clock, ''), created_at
			 FROM blob_records
			 WHERE account_id = ? AND created_at > ?
			 ORDER BY created_at ASC`,
			accountID, since,
		)
	} else {
		rows, err = d.conn.QueryContext(ctx,
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
func (d *DB) GetBlobIDs(ctx context.Context, accountID string) (map[string]bool, error) {
	rows, err := d.conn.QueryContext(ctx, `SELECT id FROM blob_records WHERE account_id = ?`, accountID)
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
