-- +goose Up
CREATE TYPE deletion_mode AS ENUM ('self', 'all');

CREATE TABLE IF NOT EXISTS messages
(
    id          UUID PRIMARY KEY,
    chat_uuid   UUID REFERENCES chats(uuid),
    sender_uuid UUID,
    content     TEXT NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    edited_at   TIMESTAMP,
    deleted_for deletion_mode
);

-- +goose Down
DROP TABLE IF EXISTS messages;
DROP TYPE IF EXISTS deletion_mode;
