-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE conv_turns ADD COLUMN "middleware_results" jsonb DEFAULT '[]';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE conv_turns DROP COLUMN IF EXISTS "middleware_results";
-- +goose StatementEnd
