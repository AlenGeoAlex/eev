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
                                           revoked_at DATETIME NULL,
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
CREATE TABLE user_target_history
(
    user_id TEXT NOT NULL,
    target_email TEXT NOT NULL,
    occurences INT NOT NULL DEFAULT 1,
    starred INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY(user_id, target_email),
    FOREIGN KEY(user_id) REFERENCES user(user_id)
);
CREATE TABLE IF NOT EXISTS "shareable_files"
(
    id           TEXT                               not null
        primary key,
    share_id     TEXT                               not null
        references shareable,
    file_name    TEXT                               not null,
    content_type TEXT                               not null,
    s3_key       TEXT                               not null,
    created_at   DATETIME default (datetime('now')) not null,
    content_size float    default 0                 not null
);
CREATE INDEX idx_shareable_files_share_id
    on shareable_files (share_id);
