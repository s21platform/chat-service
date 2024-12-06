-- +goose Up
CREATE TYPE chat_type AS ENUM ('private', 'group');

CREATE TABLE IF NOT EXISTS chats (
    id INT SERIAL PRIMARY KEY,
    uuid UUID NOT NULL,
    group_name TEXT,
    type chat_type NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_message_id BIGINT,
    owner_uuid UUID
);

-- +goose Down
DROP TABLE chats;
DROP TYPE chat_type;
