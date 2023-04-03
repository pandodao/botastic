-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE conv_turns ADD COLUMN "bot_override" jsonb DEFAULT '{}';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE conv_turns DROP COLUMN IF EXISTS "bot_override";
-- +goose StatementEnd
