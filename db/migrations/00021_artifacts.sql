-- +goose Up

CREATE TABLE artifacts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    publish_id  UUID REFERENCES publishes (id) ON DELETE SET NULL,
    name        TEXT NOT NULL,
    size        BIGINT NOT NULL,
    sha256      TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT artifacts_name_is_a_path CHECK (name <> '' AND name NOT LIKE '/%')
);

CREATE INDEX artifacts_by_project ON artifacts (project_id, created_at DESC);
CREATE INDEX artifacts_by_publish ON artifacts (publish_id);

-- +goose Down
DROP TABLE artifacts;
