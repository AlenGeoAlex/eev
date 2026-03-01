CREATE TABLE schema_migrations (version uint64,dirty bool);
CREATE UNIQUE INDEX version_unique ON schema_migrations (version);
CREATE TABLE user (
                                      id TEXT NOT NULL UNIQUE,
                                      email TEXT NOT NULL,
                                      avatar_url TEXT,
                                      source TEXT NOT NULL,
                                      created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE shareable (
                                           id TEXT primary key NOT NULL UNIQUE,
                                           name TEXT NOT NULL,
                                           user_id TEXT NOT NULL,
                                           source_ip TEXT,
                                           expiry_at DATETIME NOT NULL,
                                           active_from DATETIME NOT NULL DEFAULT (datetime('now')),
                                           created_at DATETIME NOT NULL DEFAULT (datetime('now')),
                                           shareable_type TEXT NOT NULL,
                                           shareable_data TEXT NOT NULL,
                                           FOREIGN KEY(user_id) REFERENCES user(id)
    );
CREATE TABLE shareable_options (
                                                   share_id TEXT NOT NULL,
                                                   option_key TEXT NOT NULL,
                                                   value TEXT NOT NULL,
                                                   PRIMARY KEY (share_id, option_key),
                                                   FOREIGN KEY(share_id) REFERENCES shareable(id)
    );
CREATE TABLE refresh_token (
                                               jti TEXT PRIMARY KEY NOT NULL,
                                               user_id TEXT NOT NULL,
                                               expires_at DATETIME NOT NULL,
                                               revoked INTEGER NOT NULL DEFAULT 0,
                                               created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY(user_id) REFERENCES user(id)
    );
CREATE INDEX idx_refresh_user_id
    ON refresh_token(user_id);
CREATE UNIQUE INDEX idx_user_email
    ON user(email);
CREATE TABLE shareable_files (
                                 id           TEXT PRIMARY KEY NOT NULL,
                                 share_id     TEXT NOT NULL,
                                 file_name    TEXT NOT NULL,
                                 content_type TEXT NOT NULL,
                                 s3_key       TEXT NOT NULL,
                                 created_at   DATETIME NOT NULL DEFAULT (datetime('now')),
                                 FOREIGN KEY(share_id) REFERENCES shareable(id)
);
CREATE INDEX idx_shareable_files_share_id ON shareable_files(share_id);
