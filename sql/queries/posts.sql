-- name: CreatePost :one
INSERT INTO posts(
      id,
      created_at,
      updated_at,
      title,
      url,
      description,
      published_at,
      feed_id
)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
ON CONFLICT (url) DO NOTHING
RETURNING *;

-- name: GetPostsForUser :many
SELECT posts.*, feed_follows.user_id AS user_id, users.name AS user_name FROM posts
INNER JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
INNER JOIN feeds ON feeds.id = feed_follows.feed_id
INNER JOIN users ON users.id = feed_follows.user_id
WHERE feed_follows.user_id = $1
ORDER BY posts.published_at DESC
LIMIT $2;