-- +goose Up
ALTER TABLE admin_refresh_tokens ADD retry_key CHAR(64) NOT NULL DEFAULT '', ADD successor_digest CHAR(64) NOT NULL DEFAULT '';
-- +goose Down
ALTER TABLE admin_refresh_tokens DROP COLUMN successor_digest, DROP COLUMN retry_key;
