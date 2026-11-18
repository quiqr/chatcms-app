-- +goose Up

CREATE TABLE workspaces (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    person_id    UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    project_id   UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    backend      TEXT NOT NULL,
    state        TEXT NOT NULL CHECK (state IN ('starting', 'warm', 'sleeping')),
    handle       TEXT NOT NULL DEFAULT '',
    reason       TEXT NOT NULL DEFAULT '',
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (person_id, project_id)
);

CREATE INDEX workspaces_busy ON workspaces (state) WHERE state IN ('starting', 'warm');

-- +goose Down

DROP TABLE workspaces;
