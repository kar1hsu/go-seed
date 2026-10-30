-- +goose Up
-- +goose NO TRANSACTION
CREATE TABLE IF NOT EXISTS dictionaries (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 code VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 name VARCHAR(64) NOT NULL,
 description VARCHAR(256) NOT NULL DEFAULT '',
 enabled BOOLEAN NOT NULL DEFAULT TRUE,
 created_at DATETIME(6) NOT NULL,
 updated_at DATETIME(6) NOT NULL,
 UNIQUE KEY uk_dictionary_code (code),
 KEY idx_dictionary_enabled (enabled, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE IF NOT EXISTS dictionary_items (
 dictionary_id BIGINT UNSIGNED NOT NULL,
 value VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
 label VARCHAR(64) NOT NULL,
 sort BIGINT NOT NULL DEFAULT 0,
 enabled BOOLEAN NOT NULL DEFAULT TRUE,
 PRIMARY KEY (dictionary_id, value),
 CONSTRAINT fk_dictionary_item FOREIGN KEY (dictionary_id) REFERENCES dictionaries(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE IF NOT EXISTS settings (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 code VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 name VARCHAR(64) NOT NULL,
 description VARCHAR(256) NOT NULL DEFAULT '',
 type VARCHAR(16) CHARACTER SET ascii NOT NULL,
 value TEXT NOT NULL,
 enabled BOOLEAN NOT NULL DEFAULT TRUE,
 created_at DATETIME(6) NOT NULL,
 updated_at DATETIME(6) NOT NULL,
 UNIQUE KEY uk_setting_code (code),
 KEY idx_setting_enabled (enabled, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
-- +goose Down
DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS dictionary_items;
DROP TABLE IF EXISTS dictionaries;
