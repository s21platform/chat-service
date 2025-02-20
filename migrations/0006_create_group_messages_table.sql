-- +goose Up
CREATE TABLE IF NOT EXISTS group_messages
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
    FOREIGN KEY (chat_id) REFERENCES group_chats (id),
    FOREIGN KEY (chat_id, sender_uuid) REFERENCES group_chats_user (chat_id, user_uuid),
    FOREIGN KEY (root_id) REFERENCES group_messages (id),
    FOREIGN KEY (parent_id) REFERENCES group_messages (id)
);

-- +goose Down
DROP TABLE IF EXISTS group_messages;
