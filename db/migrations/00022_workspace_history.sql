-- +goose Up

CREATE TABLE workspace_events (
    id         BIGSERIAL PRIMARY KEY,
    at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    kind       TEXT NOT NULL CHECK (kind IN ('started', 'refused', 'failed')),
    backend    TEXT NOT NULL DEFAULT '',
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    millis     INTEGER,
    reason     TEXT NOT NULL DEFAULT ''
);

CREATE INDEX workspace_events_at ON workspace_events (at);

CREATE TABLE workspace_samples (
    at       TIMESTAMPTZ PRIMARY KEY,
    starting INTEGER NOT NULL,
    warm     INTEGER NOT NULL
);

-- +goose Down

DROP TABLE workspace_samples;
DROP TABLE workspace_events;
