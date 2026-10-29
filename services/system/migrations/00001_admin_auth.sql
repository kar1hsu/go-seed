-- +goose Up
-- +goose NO TRANSACTION
CREATE TABLE IF NOT EXISTS administrators (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 username VARCHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 display_name VARCHAR(64) NOT NULL,
 password_hash VARCHAR(100) CHARACTER SET ascii NOT NULL,
 enabled BOOLEAN NOT NULL DEFAULT TRUE,
 token_version BIGINT UNSIGNED NOT NULL DEFAULT 1,
 created_at DATETIME(6) NOT NULL,
 updated_at DATETIME(6) NOT NULL,
 UNIQUE KEY uq_administrators_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE IF NOT EXISTS admin_sessions (
 id CHAR(43) CHARACTER SET ascii COLLATE ascii_bin NOT NULL PRIMARY KEY,
 admin_id BIGINT UNSIGNED NOT NULL,
 token_version BIGINT UNSIGNED NOT NULL,
 expires_at DATETIME(6) NOT NULL,
 revoked_at DATETIME(6) NULL,
 created_at DATETIME(6) NOT NULL,
 KEY idx_admin_sessions_admin (admin_id),
 CONSTRAINT fk_admin_sessions_admin FOREIGN KEY (admin_id) REFERENCES administrators(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE IF NOT EXISTS admin_refresh_tokens (
 digest CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL PRIMARY KEY,
 session_id CHAR(43) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 consumed_at DATETIME(6) NULL,
 created_at DATETIME(6) NOT NULL,
 KEY idx_refresh_session (session_id),
 CONSTRAINT fk_refresh_session FOREIGN KEY (session_id) REFERENCES admin_sessions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
-- +goose Down
DROP TABLE IF EXISTS admin_refresh_tokens;
DROP TABLE IF EXISTS admin_sessions;
DROP TABLE IF EXISTS administrators;
