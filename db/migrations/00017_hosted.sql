-- +goose Up

ALTER TABLE projects ADD COLUMN hosted BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down

ALTER TABLE projects DROP COLUMN hosted;
