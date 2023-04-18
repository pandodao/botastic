-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE bots ADD COLUMN "boundary_prompt" text DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE bots DROP COLUMN IF EXISTS "boundary_prompt";
-- +goose StatementEnd
