-- +goose Up
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS repositories (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    url        TEXT        NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    indexed_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS code_chunks (
    id         BIGSERIAL   PRIMARY KEY,
    repo_id    UUID        NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    file_path  TEXT        NOT NULL,
    content    TEXT        NOT NULL,
    embedding  vector(384),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_code_chunks_repo_id
    ON code_chunks (repo_id);

-- +goose Down
DROP INDEX IF EXISTS idx_code_chunks_repo_id;
DROP TABLE IF EXISTS code_chunks;
DROP TABLE IF EXISTS repositories;
DROP EXTENSION IF EXISTS vector;
