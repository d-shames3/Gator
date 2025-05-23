-- name: CreateFeedFollows :one
WITH inserted_feed_follows AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    )
    RETURNING *
)
SELECT
    inserted_feed_follows.id,
    inserted_feed_follows.created_at,
    inserted_feed_follows.updated_at,
    inserted_feed_follows.user_id,
    users.name as user,
    inserted_feed_follows.feed_id,
    feeds.name as feed
FROM inserted_feed_follows
INNER JOIN users
    ON inserted_feed_follows.user_id = users.id
INNER JOIN feeds
    ON inserted_feed_follows.feed_id = feeds.id
;

-- name: GetFeedFollowsForUser :many

SELECT
    feed_follows.id,
    users.name as user,
    feeds.name as feed
FROM feed_follows
INNER JOIN users
    ON feed_follows.user_id = users.id
INNER JOIN feeds
    ON feed_follows.feed_id = feeds.id
WHERE feed_follows.user_id = $1
;

-- name: DeleteFeedFollowForUser :exec
DELETE FROM feed_follows
WHERE user_id = $1 and feed_id = $2;
