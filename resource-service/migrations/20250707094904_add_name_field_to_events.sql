-- +goose Up
-- +goose StatementBegin
ALTER TABLE events ADD COLUMN name VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events DROP COLUMN name;
-- +goose StatementEnd
