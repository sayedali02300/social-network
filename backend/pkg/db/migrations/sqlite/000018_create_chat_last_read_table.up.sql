CREATE TABLE chat_last_read (
    user_id      TEXT NOT NULL,
    chat_type    TEXT NOT NULL CHECK (chat_type IN ('private', 'group')),
    chat_id      TEXT NOT NULL,
    last_read_at DATETIME NOT NULL,
    PRIMARY KEY (user_id, chat_type, chat_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
