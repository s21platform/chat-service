-- +goose Up
CREATE TABLE IF NOT EXISTS messages
(
    id          UUID PRIMARY KEY,
    chat_uuid   UUID UNIQUE,
    sender_uuid UUID,
    content     TEXT NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    edited_at   TIMESTAMP DEFAULT NULL,
    deleted     BOOLEAN   DEFAULT FALSE,
    CONSTRAINT fk_messages_chat_uuid FOREIGN KEY (chat_uuid) REFERENCES chats (uuid)
);

-- +goose Down
DROP TABLE IF EXISTS messages;
