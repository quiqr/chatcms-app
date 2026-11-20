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

-- name: BeginWorkspaceStart :one
INSERT INTO workspace_events (at, kind, backend, project_id, person_id)
VALUES ($1, 'starting', $2, $3, $4)
RETURNING id;

-- name: FinishWorkspaceStart :exec
UPDATE workspace_events
SET kind = $2,
    millis = $3,
    reason = $4,
    steps = $5,
    log = $6,
    hugo = $7,
    finished_at = $8
WHERE id = $1;

-- name: InterruptWorkspaceStarts :exec
UPDATE workspace_events
SET kind = 'interrupted', finished_at = now()
WHERE kind = 'starting';

-- name: GetWorkspaceStart :one
SELECT e.id, e.at, e.kind, e.backend, e.project_id, e.millis, e.reason, e.steps, e.log, e.hugo, e.finished_at,
       coalesce(nullif(p.display_name, ''), p.email, '')::text AS who
FROM workspace_events e
LEFT JOIN people p ON p.id = e.person_id
WHERE e.id = $1 AND e.project_id = $2 AND e.kind <> 'refused';

-- name: ProjectLog :many
SELECT kind, id, at, finished_at, state, what, who, millis FROM (
    SELECT 'start'::text AS kind, e.id::text AS id, e.at AS at, e.finished_at AS finished_at,
           e.kind::text AS state, e.backend::text AS what,
           coalesce(nullif(p.display_name, ''), p.email, '')::text AS who,
           e.millis AS millis
    FROM workspace_events e
    LEFT JOIN people p ON p.id = e.person_id
    WHERE e.project_id = sqlc.arg(project) AND e.kind <> 'refused'
    UNION ALL
    SELECT 'build'::text, b.id::text, b.started_at, b.finished_at,
           b.state::text, b.reason::text, ''::text,
           (extract(epoch FROM b.finished_at - b.started_at) * 1000)::int
    FROM builds b
    WHERE b.project_id = sqlc.arg(project)
    UNION ALL
    SELECT 'publish'::text, u.id::text, u.started_at, u.finished_at,
           u.state::text, u.change_count::text,
           coalesce(nullif(p.display_name, ''), p.email, '')::text,
           (extract(epoch FROM u.finished_at - u.started_at) * 1000)::int
    FROM publishes u
    LEFT JOIN people p ON p.id = u.person_id
    WHERE u.project_id = sqlc.arg(project)
) entries
WHERE (sqlc.narg(of_kind)::text IS NULL OR kind = sqlc.narg(of_kind)::text)
  AND (sqlc.narg(before)::timestamptz IS NULL OR at < sqlc.narg(before)::timestamptz)
ORDER BY at DESC, id DESC
LIMIT sqlc.arg(most);
