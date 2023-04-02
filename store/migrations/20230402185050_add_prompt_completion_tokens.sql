-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE conv_turns ADD COLUMN "prompt_tokens" int DEFAULT 0;
ALTER TABLE conv_turns ADD COLUMN "completion_tokens" int DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE conv_turns DROP COLUMN IF EXISTS "prompt_tokens";
ALTER TABLE conv_turns DROP COLUMN IF EXISTS "completion_tokens";
-- +goose StatementEnd
