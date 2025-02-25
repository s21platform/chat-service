-- +goose Up
ALTER TABLE chats
    ADD CONSTRAINT fk_chats_deleted_by
        FOREIGN KEY (uuid, deleted_by) REFERENCES chats_user (chat_uuid, user_uuid);

ALTER TABLE group_chats
    ADD CONSTRAINT fk_group_chats_deleted_by
        FOREIGN KEY (uuid, deleted_by) REFERENCES group_chats_user (chat_uuid, user_uuid);

-- +goose Down
ALTER TABLE chats
    DROP CONSTRAINT fk_chats_deleted_by;

ALTER TABLE group_chats
    DROP CONSTRAINT fk_group_chats_deleted_by;
