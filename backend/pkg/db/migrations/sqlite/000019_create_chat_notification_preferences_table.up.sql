CREATE TABLE chat_notification_preferences (
    user_id   TEXT NOT NULL,
    chat_type TEXT NOT NULL CHECK (chat_type IN ('private', 'group')),
    chat_id   TEXT NOT NULL,
    muted     INTEGER NOT NULL DEFAULT 0 CHECK (muted IN (0, 1)),
    PRIMARY KEY (user_id, chat_type, chat_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
