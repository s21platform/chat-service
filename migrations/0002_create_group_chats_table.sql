-- +goose Up
CREATE TABLE IF NOT EXISTS group_chats
(
    uuid         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    deleted_by UUID
);

-- +goose Down
DROP TABLE IF EXISTS group_chats;
