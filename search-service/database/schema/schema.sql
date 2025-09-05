CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    topic VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    sent BOOLEAN NOT NULL DEFAULT FALSE,
    event_time TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index on sent column for efficient querying of unsent events
CREATE INDEX IF NOT EXISTS idx_events_sent ON events (sent);

-- Index on event_time for chronological processing
CREATE INDEX IF NOT EXISTS idx_events_event_time ON events (event_time);