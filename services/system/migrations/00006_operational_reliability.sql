-- +goose NO TRANSACTION
-- +goose Up
ALTER TABLE audit_logs MODIFY target_name VARCHAR(180) NOT NULL;
ALTER TABLE task_outbox ADD delivery BIGINT UNSIGNED NOT NULL DEFAULT 0;
ALTER TABLE tasks ADD KEY idx_task_file (file_id);
ALTER TABLE files ADD cleanup_after DATETIME(6) NULL;

-- +goose Down
ALTER TABLE files DROP COLUMN cleanup_after;
ALTER TABLE tasks DROP KEY idx_task_file;
ALTER TABLE task_outbox DROP COLUMN delivery;
ALTER TABLE audit_logs MODIFY target_name VARCHAR(64) NOT NULL;
