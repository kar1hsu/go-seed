-- +goose Up
CREATE TABLE menu_layouts (
    id TINYINT UNSIGNED PRIMARY KEY,
    revision BIGINT UNSIGNED NOT NULL,
    layout JSON NOT NULL
) ENGINE=InnoDB;
INSERT INTO menu_layouts (id, revision, layout) VALUES (1, 1, JSON_ARRAY());

-- +goose Down
DROP TABLE menu_layouts;
