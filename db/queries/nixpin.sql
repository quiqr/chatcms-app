-- name: GetNixPackage :one
SELECT * FROM nix_packages WHERE name = $1 AND version = $2 AND system = $3;

-- name: PutNixPackage :exec
INSERT INTO nix_packages (name, version, system, revision, attr_path, store_path, realised_at)
VALUES ($1, $2, $3, $4, $5, $6, now())
ON CONFLICT (name, version, system) DO UPDATE
SET revision = EXCLUDED.revision,
    attr_path = EXCLUDED.attr_path,
    store_path = EXCLUDED.store_path,
    realised_at = now();
