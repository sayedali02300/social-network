CREATE TABLE IF NOT EXISTS post_likes (
    post_id    TEXT    NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id    TEXT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    value      INTEGER NOT NULL CHECK (value IN (1, -1)),
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (post_id, user_id)
);
