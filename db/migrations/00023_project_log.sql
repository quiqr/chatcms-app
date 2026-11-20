-- +goose Up

ALTER TABLE workspace_events DROP CONSTRAINT workspace_events_kind_check;
ALTER TABLE workspace_events ADD CONSTRAINT workspace_events_kind_check
    CHECK (kind IN ('starting', 'started', 'refused', 'failed', 'interrupted'));
ALTER TABLE workspace_events
    ADD COLUMN person_id   UUID REFERENCES people(id) ON DELETE SET NULL,
    ADD COLUMN steps       JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN log         TEXT NOT NULL DEFAULT '',
    ADD COLUMN hugo        JSONB,
    ADD COLUMN finished_at TIMESTAMPTZ;

CREATE INDEX workspace_events_by_project ON workspace_events (project_id, at DESC);

-- +goose Down

DROP INDEX workspace_events_by_project;
ALTER TABLE workspace_events
    DROP COLUMN finished_at,
    DROP COLUMN hugo,
    DROP COLUMN log,
    DROP COLUMN steps,
    DROP COLUMN person_id;
DELETE FROM workspace_events WHERE kind IN ('starting', 'interrupted');
ALTER TABLE workspace_events DROP CONSTRAINT workspace_events_kind_check;
ALTER TABLE workspace_events ADD CONSTRAINT workspace_events_kind_check
    CHECK (kind IN ('started', 'refused', 'failed'));
