-- +goose Up

CREATE TABLE template_feed (
    id         BOOLEAN     PRIMARY KEY DEFAULT TRUE CHECK (id),
    feed_url   TEXT        NOT NULL,
    templates  JSONB       NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE projects
    ADD COLUMN hugo_version  TEXT    NOT NULL DEFAULT '',
    ADD COLUMN hugo_extended BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down

ALTER TABLE projects DROP COLUMN hugo_extended, DROP COLUMN hugo_version;

DROP TABLE template_feed;
