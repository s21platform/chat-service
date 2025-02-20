-- +goose Up
CREATE TABLE IF NOT EXISTS chats_user
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id     UUID NOT NULL,
    user_uuid   UUID NOT NULL,
    username    TEXT NOT NULL,
    avatar_link TEXT NOT NULL,
    joined_at   TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (chat_id) REFERENCES chats (id),
    CONSTRAINT unique_chat_user UNIQUE (chat_id, user_uuid)
);

-- +goose Down
DROP TABLE IF EXISTS chats_user;
