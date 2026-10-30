-- +goose Up
-- +goose NO TRANSACTION
CREATE TABLE IF NOT EXISTS audit_logs (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 action VARCHAR(96) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 actor_id BIGINT UNSIGNED NOT NULL,
 actor_username VARCHAR(32) NOT NULL,
 actor_name VARCHAR(64) NOT NULL,
 source VARCHAR(16) CHARACTER SET ascii NOT NULL,
 target_type VARCHAR(32) CHARACTER SET ascii NOT NULL,
 target_id BIGINT UNSIGNED NOT NULL,
 target_key VARCHAR(64) NOT NULL,
 target_name VARCHAR(64) NOT NULL,
 trace_id CHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
 changes JSON NOT NULL,
 created_at DATETIME(6) NOT NULL,
 KEY idx_audit_time (created_at, id),
 KEY idx_audit_actor (actor_id, created_at, id),
 KEY idx_audit_action (action, created_at, id),
 KEY idx_audit_trace (trace_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
-- Historical actor and target snapshots deliberately have no cascading foreign keys.
-- +goose Down
DROP TABLE IF EXISTS audit_logs;
