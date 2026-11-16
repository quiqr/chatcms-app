-- name: SaveTemplateFeed :exec
INSERT INTO template_feed (id, feed_url, templates, fetched_at)
VALUES (TRUE, $1, $2, now())
ON CONFLICT (id) DO UPDATE
SET feed_url = EXCLUDED.feed_url, templates = EXCLUDED.templates, fetched_at = EXCLUDED.fetched_at;

-- name: GetTemplateFeed :one
SELECT feed_url, templates, fetched_at FROM template_feed WHERE id;
