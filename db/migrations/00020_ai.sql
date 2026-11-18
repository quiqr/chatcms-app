-- +goose Up

CREATE TABLE ai_tokens (
    token_hash      BYTEA       PRIMARY KEY,
    person_id       UUID        NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    project_id      UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    organisation_id UUID        NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    issued_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at      TIMESTAMPTZ
);

CREATE INDEX ai_tokens_live ON ai_tokens (person_id, project_id) WHERE revoked_at IS NULL;

CREATE TABLE ai_usage (
    id              BIGSERIAL   PRIMARY KEY,
    at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    organisation_id UUID        NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    project_id      UUID        REFERENCES projects(id) ON DELETE SET NULL,
    person_id       UUID        REFERENCES people(id) ON DELETE SET NULL,
    model           TEXT        NOT NULL,
    tokens_in       BIGINT      NOT NULL DEFAULT 0,
    tokens_out      BIGINT      NOT NULL DEFAULT 0,
    status          INTEGER     NOT NULL
);

CREATE INDEX ai_usage_by_organisation ON ai_usage (organisation_id, at);

ALTER TABLE organisations
    ADD COLUMN ai_requests_per_minute INTEGER,
    ADD COLUMN ai_tokens_per_month    BIGINT;

ALTER TABLE organisations
    ADD CONSTRAINT organisations_ai_requests_positive
        CHECK (ai_requests_per_minute IS NULL OR ai_requests_per_minute > 0),
    ADD CONSTRAINT organisations_ai_tokens_positive
        CHECK (ai_tokens_per_month IS NULL OR ai_tokens_per_month > 0);

-- +goose Down

ALTER TABLE organisations
    DROP COLUMN ai_requests_per_minute,
    DROP COLUMN ai_tokens_per_month;

DROP TABLE ai_usage;
DROP TABLE ai_tokens;
