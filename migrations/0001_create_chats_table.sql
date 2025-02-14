-- +goose Up
CREATE TYPE chat_type AS ENUM ('private', 'group');

CREATE TABLE IF NOT EXISTS chats
(
    id              SERIAL PRIMARY KEY,
    uuid            UUID UNIQUE NOT NULL,
    chat_name      TEXT,
    type            chat_type   NOT NULL,
    avatar_link     TEXT NOT NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_message_id UUID,
    owner_uuid      UUID
);

-- +goose Down
DROP TABLE IF EXISTS chats;
DROP TYPE IF EXISTS chat_type;
