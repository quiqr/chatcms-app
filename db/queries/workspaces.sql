-- name: LockWorkspaces :exec
SELECT pg_advisory_xact_lock(7307);

-- name: CountBusyWorkspacesBesides :one
SELECT count(*) FROM workspaces
WHERE state IN ('starting', 'warm')
  AND NOT (person_id = $1 AND project_id = $2);

-- name: StartWorkspace :one
INSERT INTO workspaces (person_id, project_id, backend, state, last_used_at)
VALUES ($1, $2, $3, 'starting', $4)
ON CONFLICT (person_id, project_id) DO UPDATE
SET backend = EXCLUDED.backend,
    state = 'starting',
    handle = '',
    reason = '',
    last_used_at = EXCLUDED.last_used_at
RETURNING *;

-- name: WarmWorkspace :exec
UPDATE workspaces SET state = 'warm', handle = $2, reason = '' WHERE id = $1;

-- name: SleepWorkspace :exec
UPDATE workspaces SET state = 'sleeping', handle = '', reason = $2 WHERE id = $1;

-- name: TouchWorkspace :exec
UPDATE workspaces SET last_used_at = $2 WHERE id = $1 AND last_used_at < $2;

-- name: ListBusyWorkspaces :many
SELECT * FROM workspaces WHERE state IN ('starting', 'warm') ORDER BY created_at;

-- name: ListSleepingWorkspaces :many
SELECT * FROM workspaces WHERE state = 'sleeping' ORDER BY created_at;

-- name: GetWorkspace :one
SELECT * FROM workspaces WHERE person_id = $1 AND project_id = $2;

-- name: DeleteWorkspace :exec
DELETE FROM workspaces WHERE id = $1;

-- name: RecordWorkspaceEvent :exec
INSERT INTO workspace_events (at, kind, backend, project_id, millis, reason)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: SampleWorkspaces :exec
INSERT INTO workspace_samples (at, starting, warm) VALUES ($1, $2, $3)
ON CONFLICT (at) DO UPDATE SET starting = EXCLUDED.starting, warm = EXCLUDED.warm;

-- name: ForgetWorkspaceEvents :exec
DELETE FROM workspace_events WHERE at < $1;

-- name: ForgetWorkspaceSamples :exec
DELETE FROM workspace_samples WHERE at < $1;

-- name: CountWorkspacesByState :many
SELECT w.state, w.backend, o.slug AS organisation_slug, o.name AS organisation_name,
       p.id AS project_id, p.slug AS project_slug, p.name AS project_name, count(*)::int AS workspaces
FROM workspaces w
JOIN projects p ON p.id = w.project_id
JOIN organisations o ON o.id = p.organisation_id
GROUP BY w.state, w.backend, o.slug, o.name, p.id, p.slug, p.name
ORDER BY o.slug, p.slug, w.state;

-- name: WorkspaceStartsSince :one
SELECT count(*)::int AS starts,
       coalesce(percentile_cont(0.5) WITHIN GROUP (ORDER BY millis), 0)::float8 AS median_millis,
       coalesce(percentile_cont(0.9) WITHIN GROUP (ORDER BY millis), 0)::float8 AS slow_millis
FROM workspace_events
WHERE kind = 'started' AND at >= $1;

-- name: CountWorkspaceEventsSince :many
SELECT kind, reason, count(*)::int AS times
FROM workspace_events
WHERE kind IN ('refused', 'failed') AND at >= $1
GROUP BY kind, reason
ORDER BY kind, times DESC, reason;

-- name: WorkspacePeaksSince :many
SELECT date_trunc('day', at)::date AS day, max(warm)::int AS warm, max(starting)::int AS starting
FROM workspace_samples
WHERE at >= $1
GROUP BY 1
ORDER BY 1;
