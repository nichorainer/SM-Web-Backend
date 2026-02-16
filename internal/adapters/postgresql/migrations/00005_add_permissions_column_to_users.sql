-- +goose Up
-- +goose StatementBegin

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS permissions TEXT[] DEFAULT '{}';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE users
  DROP COLUMN IF EXISTS permissions;

-- +goose StatementEnd