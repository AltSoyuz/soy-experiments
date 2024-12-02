CREATE TABLE IF NOT EXISTS user (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	email TEXT NOT NULL,
	password_hash TEXT NOT NULL,
    email_verified INTEGER NOT NULL DEFAULT 0,
	created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS session (
    id TEXT NOT NULL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES user(id),
    expires_at INTEGER NOT NULL,
	created_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS email_verification_request (
  	user_id INTEGER NOT NULL UNIQUE PRIMARY KEY REFERENCES user(id),
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    code TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS password_reset_request (
	id INTEGER NOT NULL UNIQUE PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL UNIQUE REFERENCES user(id),
    created_at TEXT DEFAULT (datetime('now')),
    expires_at INTEGER NOT NULL,
    code_hash TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS oauth_accounts (
    user_id TEXT NOT NULL UNIQUE PRIMARY KEY REFERENCES user(id),
    provider TEXT NOT NULL,
    provider_user_id TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now'))
);


CREATE TABLE IF NOT EXISTS todos (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL REFERENCES user(id),
	name TEXT NOT NULL,
	description TEXT
);
