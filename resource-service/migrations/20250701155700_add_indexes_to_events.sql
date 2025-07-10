-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_events_sent ON events USING HASH (sent);
CREATE INDEX IF NOT EXISTS idx_events_topic ON events USING HASH (topic);
CREATE INDEX IF NOT EXISTS idx_events_event_time ON events (event_time DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_events_sent;
DROP INDEX IF EXISTS idx_events_topic;
DROP INDEX IF EXISTS idx_events_event_time;
-- +goose StatementEnd
