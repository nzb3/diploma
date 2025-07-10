-- +goose Up
-- +goose StatementBegin
CREATE TABLE events (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        topic VARCHAR(255) NOT NULL,
                        payload JSON NOT NULL,
                        sent BOOLEAN NOT NULL DEFAULT FALSE,
                        event_time TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events;
-- +goose StatementEnd
