-- +goose Up
CREATE TYPE stream_type AS ENUM ('private', 'group', 'comment', 'channel');
CREATE TABLE IF NOT EXISTS streams
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type       stream_type NOT NULL,
    metadata   JSONB,
    created_at TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    FOREIGN KEY (created_by) REFERENCES users (id)
);

-- +goose Down
DROP TABLE IF EXISTS streams;
DROP TYPE IF EXISTS stream_type;
