-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE resource_type AS ENUM (
    'pdf', 'txt', 'url'
    );

CREATE TYPE resource_status AS ENUM (
    'pending', 'processing', 'completed', 'failed'
    );

CREATE TABLE IF NOT EXISTS resources (
                           id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                           name VARCHAR(255) NOT NULL,
                           type resource_type NOT NULL,
                           url VARCHAR(255),
                           extracted_content TEXT,
                           raw_content BYTEA,
                           status resource_status NOT NULL DEFAULT 'pending',
                           owner_id UUID NOT NULL,
                           created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                           updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS resources;
DROP TYPE IF EXISTS resource_status;
DROP TYPE IF EXISTS resource_type;
-- +goose StatementEnd
