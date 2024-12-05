-- +goose Up
CREATE TYPE chat_type AS ENUM ('private', 'group');

CREATE TABLE IF NOT EXISTS chats (
    id INT SERIAL PRIMARY KEY,
    uuid UUID NOT NULL,
    group_name TEXT,
    type chat_type NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_message_id BIGINT,
    owner_uuid UUID,
    CONSTRAINT fk_chats_last_message FOREIGN KEY (last_message_id) REFERENCES messages(id) ON DELETE SET NULL,
    CONSTRAINT fk_chats_owner_uuid FOREIGN KEY (owner_uuid) REFERENCES chat_members (user_uuid) ON DELETE SET NULL
);

-- +goose Down
DROP TABLE chats;