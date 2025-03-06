-- +goose Up
CREATE TABLE IF NOT EXISTS group_messages
(
    uuid            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_uuid       UUID      NOT NULL,
    sender_uuid   UUID      NOT NULL,
    content       TEXT      NOT NULL,
    sent_at       TIMESTAMP NOT NULL,
    updated_at    TIMESTAMP,
    root_uuid       UUID,
    parent_uuid     UUID,
    deleted_at    TIMESTAMP,
    delete_format delete_format_type,
    deleted_by UUID,
    FOREIGN KEY (chat_uuid) REFERENCES group_chats (uuid),
    FOREIGN KEY (chat_uuid, sender_uuid) REFERENCES group_chats_user (chat_uuid, user_uuid),
    FOREIGN KEY (root_uuid) REFERENCES group_messages (uuid),
    FOREIGN KEY (parent_uuid) REFERENCES group_messages (uuid),
    FOREIGN KEY (chat_uuid, deleted_by) REFERENCES group_chats_user(chat_uuid, user_uuid)
);

-- +goose Down
DROP TABLE IF EXISTS group_messages;
