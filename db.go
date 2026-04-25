package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func openDB(dataDir string) *sql.DB {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("failed to create data dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dataDir, "uploads"), 0755); err != nil {
		log.Fatalf("failed to create uploads dir: %v", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(dataDir, "defiver.db"))
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		log.Fatalf("failed to set WAL mode: %v", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		log.Fatalf("failed to enable foreign keys: %v", err)
	}

	migrate(db)
	return db
}

func migrate(db *sql.DB) {
	schema := `
CREATE TABLE IF NOT EXISTS agents (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	api_key_hash TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	currency_preference TEXT NOT NULL DEFAULT 'USDT',
	wallet_address TEXT NOT NULL DEFAULT '',
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	display_name TEXT NOT NULL,
	wallet_address TEXT NOT NULL DEFAULT '',
	bio TEXT NOT NULL DEFAULT '',
	reputation_score REAL NOT NULL DEFAULT 0.0,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	agent_id INTEGER NOT NULL REFERENCES agents(id),
	title TEXT NOT NULL,
	description TEXT NOT NULL,
	max_budget REAL NOT NULL,
	currency TEXT NOT NULL DEFAULT 'USDT',
	deadline DATETIME NOT NULL,
	status TEXT NOT NULL DEFAULT 'open',
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS bids (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id INTEGER NOT NULL REFERENCES tasks(id),
	user_id INTEGER NOT NULL REFERENCES users(id),
	amount REAL NOT NULL,
	delivery_days INTEGER NOT NULL,
	message TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'pending',
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS deliveries (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id INTEGER NOT NULL REFERENCES tasks(id),
	bid_id INTEGER NOT NULL REFERENCES bids(id),
	user_id INTEGER NOT NULL REFERENCES users(id),
	content_text TEXT NOT NULL DEFAULT '',
	file_path TEXT NOT NULL DEFAULT '',
	file_name TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'pending',
	submitted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	reviewed_at DATETIME
);

CREATE TABLE IF NOT EXISTS payments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id INTEGER NOT NULL REFERENCES tasks(id),
	bid_id INTEGER NOT NULL REFERENCES bids(id),
	amount REAL NOT NULL,
	currency TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'pending',
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`
	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}
