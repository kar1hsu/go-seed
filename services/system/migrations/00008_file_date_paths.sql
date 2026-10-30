-- +goose NO TRANSACTION
-- +goose Up
ALTER TABLE files MODIFY object_key VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL;

-- +goose Down
-- Dated keys cannot be truncated without losing the location of stored files.
SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'File date paths require VARCHAR(64); automatic rollback would truncate object keys';
