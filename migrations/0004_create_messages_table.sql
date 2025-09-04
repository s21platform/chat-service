-- +goose Up
CREATE TYPE message_type AS ENUM ('text', 'image', 'video', 'file', 'speech', 'circle');
CREATE TYPE delete_format_type AS ENUM ('self', 'all');
CREATE TABLE IF NOT EXISTS messages
(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id     UUID NOT NULL,
    sender_id     UUID NOT NULL,
    type          message_type,
    content       TEXT,
    media         JSONB,
    root_id       UUID,
    parent_id     UUID,
    sent_at       TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP,
    deleted_at    TIMESTAMP,
    delete_format delete_format_type,
    deleted_by    UUID,
    FOREIGN KEY (stream_id) REFERENCES streams (id),
    FOREIGN KEY (stream_id, sender_id) REFERENCES stream_members (stream_id, user_id),
    FOREIGN KEY (root_id) REFERENCES messages (id),
    FOREIGN KEY (parent_id) REFERENCES messages (id),
    FOREIGN KEY (stream_id, deleted_by) REFERENCES stream_members (stream_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS messages;
DROP TYPE IF EXISTS delete_format_type;
DROP TYPE IF EXISTS message_type;
