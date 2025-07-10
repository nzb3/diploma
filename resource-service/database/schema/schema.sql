CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE resource_type AS ENUM (
    'pdf', 'txt', 'url'
    );

CREATE TYPE resource_status AS ENUM (
    'pending', 'processing', 'completed', 'failed'
    );

CREATE TABLE resources (
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

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    topic VARCHAR(255) NOT NULL,
    payload JSON NOT NULL,
    sent BOOLEAN NOT NULL DEFAULT FALSE,
    event_time TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_resources_status ON resources USING HASH (status);
CREATE INDEX IF NOT EXISTS idx_resources_type ON resources USING HASH (type);
CREATE INDEX IF NOT EXISTS idx_resources_owner_id ON resources (owner_id);
CREATE INDEX IF NOT EXISTS idx_resources_created_at ON resources (created_at DESC);
