-- +goose Up
CREATE TABLE IF NOT EXISTS stream_members
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id  UUID  NOT NULL,
    user_id    UUID  NOT NULL,
    metadata   JSONB NOT NULL,
    notify     BOOLEAN          DEFAULT TRUE,
    banned_at  TIMESTAMP,
    banned_by  UUID,
    ban_reason TEXT,
    ban_until  TIMESTAMP,
    joined_at  TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    invited_by UUID,
    left_at    TIMESTAMP,
    FOREIGN KEY (stream_id) REFERENCES streams (id),
    FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT unique_stream_user UNIQUE (stream_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS stream_members;
