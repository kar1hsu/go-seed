-- +goose NO TRANSACTION
-- +goose Up
CREATE TABLE IF NOT EXISTS files (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 owner_id BIGINT UNSIGNED NOT NULL,
 name VARCHAR(180) NOT NULL,
 object_key CHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 size BIGINT NOT NULL,
 content_type VARCHAR(80) NOT NULL,
 status VARCHAR(16) NOT NULL,
 source VARCHAR(16) NOT NULL,
 sha256 CHAR(64) CHARACTER SET ascii NOT NULL DEFAULT '',
 created_at DATETIME(6) NOT NULL,
 updated_at DATETIME(6) NOT NULL,
 UNIQUE KEY uk_file_object (object_key),
 KEY idx_file_owner (owner_id,id),
 KEY idx_file_cleanup (status,created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE IF NOT EXISTS tasks (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 owner_id BIGINT UNSIGNED NOT NULL,
 kind VARCHAR(64) CHARACTER SET ascii NOT NULL,
 dedup_key VARCHAR(128) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 status VARCHAR(16) NOT NULL,
 generation BIGINT NOT NULL DEFAULT 1,
 attempts BIGINT NOT NULL DEFAULT 0,
 max_attempts BIGINT NOT NULL DEFAULT 5,
 file_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
 error VARCHAR(256) NOT NULL DEFAULT '',
 trace_id CHAR(32) CHARACTER SET ascii NOT NULL DEFAULT '',
 lease_until DATETIME(6) NULL,
 next_run_at DATETIME(6) NOT NULL,
 created_at DATETIME(6) NOT NULL,
 updated_at DATETIME(6) NOT NULL,
 UNIQUE KEY uk_task_dedup (dedup_key),
 KEY idx_task_owner (owner_id,id),
 KEY idx_task_recover (status,next_run_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE IF NOT EXISTS task_outbox (
 task_id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
 generation BIGINT NOT NULL,
 available_at DATETIME(6) NOT NULL,
 delivered_at DATETIME(6) NULL,
 KEY idx_outbox_available (available_at),
 CONSTRAINT fk_outbox_task FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
-- +goose Down
DROP TABLE IF EXISTS task_outbox;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS files;
