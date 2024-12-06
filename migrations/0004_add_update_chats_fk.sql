-- +goose Up
ALTER TABLE chats
ADD CONSTRAINT fk_chats_last_message FOREIGN KEY (last_message_id) REFERENCES messages(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE chats
DROP CONSTRAINT fk_chats_last_message;
