-- +goose Up
CREATE TABLE feed_follows(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    feed_id UUID NOT NULL,
    user_id UUID NOT NULL,

    CONSTRAINT no_orphaned_followers
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT no_orphaned_feeds
        FOREIGN KEY (feed_id)
        REFERENCES feeds(id)
        ON DELETE CASCADE,

    CONSTRAINT no_duplicate_relations UNIQUE (user_id, feed_id)
);

-- +goose Down
DROP TABLE feed_follows;