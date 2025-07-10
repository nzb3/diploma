-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_resources_status ON resources USING HASH (status);
CREATE INDEX IF NOT EXISTS idx_resources_type ON resources USING HASH (type);
CREATE INDEX IF NOT EXISTS idx_resources_owner_id ON resources (owner_id);
CREATE INDEX IF NOT EXISTS idx_resources_created_at ON resources (created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_resources_created_at;
DROP INDEX IF EXISTS idx_resources_owner_id;
DROP INDEX IF EXISTS idx_resources_type;
DROP INDEX IF EXISTS idx_resources_status;
-- +goose StatementEnd
