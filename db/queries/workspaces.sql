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
