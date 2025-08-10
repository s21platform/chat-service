-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE TABLE IF NOT EXISTS users
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nickname    TEXT NOT NULL,
    avatar_url  TEXT NOT NULL,
    last_online TIMESTAMP,
    created_at  TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS pgcrypto;
