-- +goose Up
CREATE TABLE IF NOT EXISTS group_chats
(
    uuid         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    chat_name TEXT NOT NULL,
    avatar_link TEXT NOT NULL,
    deleted_at TIMESTAMP,
    deleted_by UUID
);

-- +goose Down
DROP TABLE IF EXISTS group_chats;
