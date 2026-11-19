-- name: InsertArtifact :exec
INSERT INTO artifacts (id, project_id, publish_id, name, size, sha256)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListArtifacts :many
SELECT * FROM artifacts WHERE project_id = $1 ORDER BY created_at DESC, name LIMIT $2;

-- name: GetArtifact :one
SELECT * FROM artifacts WHERE project_id = $1 AND id = $2;

-- name: ArtifactsOfPublish :many
SELECT * FROM artifacts WHERE publish_id = $1 ORDER BY name;

-- name: ProjectHasArtifacts :one
SELECT EXISTS (SELECT 1 FROM artifacts WHERE project_id = $1);
