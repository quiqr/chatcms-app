-- name: RevokeAITokensFor :exec
UPDATE ai_tokens SET revoked_at = now()
WHERE person_id = $1 AND project_id = $2 AND revoked_at IS NULL;

-- name: IssueAIToken :exec
INSERT INTO ai_tokens (token_hash, person_id, project_id, organisation_id)
SELECT $1, $2, $3, p.organisation_id FROM projects p WHERE p.id = $3;

-- name: LookUpAIToken :one
SELECT person_id, project_id, organisation_id FROM ai_tokens
WHERE token_hash = $1 AND revoked_at IS NULL;

-- name: RecordAIUsage :exec
INSERT INTO ai_usage (organisation_id, project_id, person_id, model, tokens_in, tokens_out, status)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: AIUsageSince :one
SELECT count(*)::BIGINT AS requests,
       COALESCE(sum(tokens_in + tokens_out), 0)::BIGINT AS tokens
FROM ai_usage
WHERE organisation_id = $1 AND at >= $2;

-- name: AILimitsOf :one
SELECT ai_requests_per_minute, ai_tokens_per_month FROM organisations WHERE id = $1;

-- name: SetAILimits :execrows
UPDATE organisations
SET ai_requests_per_minute = $2, ai_tokens_per_month = $3, updated_at = now()
WHERE id = $1;
