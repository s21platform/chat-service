-- +goose Up
CREATE TYPE delete_format_type AS ENUM ('self', 'all');

CREATE TABLE IF NOT EXISTS messages
(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id       UUID      NOT NULL,
    sender_uuid   UUID      NOT NULL,
    content       TEXT      NOT NULL,
    sent_at       TIMESTAMP NOT NULL,
    updated_at    TIMESTAMP,
    root_id       UUID,
    parent_id     UUID,
    deleted_at    TIMESTAMP,
    delete_format delete_format_type,
    FOREIGN KEY (chat_id) REFERENCES chats (id),
    FOREIGN KEY (chat_id, sender_uuid) REFERENCES chats_user (chat_id, user_uuid),
    FOREIGN KEY (root_id) REFERENCES messages (id),
    FOREIGN KEY (parent_id) REFERENCES messages (id)
);

-- +goose Down
DROP TABLE IF EXISTS messages;
DROP TYPE IF EXISTS delete_format_type;
