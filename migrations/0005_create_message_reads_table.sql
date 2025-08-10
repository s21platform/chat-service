-- +goose Up
CREATE TABLE IF NOT EXISTS message_reads
(
    id         UUID PRIMARY KEY   DEFAULT gen_random_uuid(),
    message_id UUID      NOT NULL,
    user_id    UUID      NOT NULL,
    read_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES messages (id),
    FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT unique_message_user_read UNIQUE (message_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS message_reads;
