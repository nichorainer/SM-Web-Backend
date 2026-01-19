-- +goose Up
-- +goose StatementBegin
-- 00004_seed_first_user_as_admin.sql
-- This migration ensures there is at least one admin.
-- It will promote the earliest user to admin if no admin exists.
-- It will NOT create a seeded account with a known password.
DO $$
BEGIN
  -- If there is no admin, promote the earliest-registered user to admin
  IF (SELECT COUNT(*) FROM users WHERE role = 'admin') = 0 THEN
    WITH first_user AS (
      SELECT id FROM users ORDER BY created_at ASC LIMIT 1
    )
    UPDATE users
    SET role = 'admin'
    FROM first_user
    WHERE users.id = first_user.id;
  END IF;
END$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- No down migration needed for seeding data
-- +goose StatementEnd
