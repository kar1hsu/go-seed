-- +goose Up
-- +goose NO TRANSACTION
CREATE TABLE IF NOT EXISTS users (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 username VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 nickname VARCHAR(64) NOT NULL,
 password_hash VARCHAR(100) CHARACTER SET ascii NOT NULL,
 enabled BOOLEAN NOT NULL DEFAULT TRUE,
 token_version BIGINT UNSIGNED NOT NULL DEFAULT 1,
 created_at DATETIME(6) NOT NULL,
 updated_at DATETIME(6) NOT NULL,
 last_login_at DATETIME(6) NULL,
 UNIQUE KEY uq_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE IF NOT EXISTS member_sessions (
 id CHAR(43) CHARACTER SET ascii COLLATE ascii_bin NOT NULL PRIMARY KEY,
 user_id BIGINT UNSIGNED NOT NULL,
 token_version BIGINT UNSIGNED NOT NULL,
 expires_at DATETIME(6) NOT NULL,
 revoked_at DATETIME(6) NULL,
 created_at DATETIME(6) NOT NULL,
 KEY idx_member_sessions_admin (user_id),
 CONSTRAINT fk_member_sessions_admin FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE IF NOT EXISTS member_refresh_tokens (
 digest CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL PRIMARY KEY,
 session_id CHAR(43) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 consumed_at DATETIME(6) NULL,
 created_at DATETIME(6) NOT NULL,
 KEY idx_refresh_session (session_id),
 CONSTRAINT fk_refresh_session FOREIGN KEY (session_id) REFERENCES member_sessions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE IF NOT EXISTS audit_events (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 user_id BIGINT UNSIGNED NOT NULL,
 action VARCHAR(32) CHARACTER SET ascii NOT NULL,
 actor_kind VARCHAR(16) CHARACTER SET ascii NOT NULL,
 actor_id BIGINT UNSIGNED NOT NULL,
 trace_id VARCHAR(64) CHARACTER SET ascii NOT NULL,
 created_at DATETIME(6) NOT NULL,
 KEY idx_audit_user (user_id,id),
 CONSTRAINT fk_audit_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
-- +goose Down
DROP TABLE IF EXISTS audit_events;
DROP TABLE IF EXISTS member_refresh_tokens;
DROP TABLE IF EXISTS member_sessions;
DROP TABLE IF EXISTS users;
