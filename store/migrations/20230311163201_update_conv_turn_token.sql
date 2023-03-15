-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE conv_turns DROP COLUMN IF EXISTS "request_token";
ALTER TABLE conv_turns RENAME COLUMN "response_token" TO "total_tokens";

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE conv_turns RENAME COLUMN "total_tokens" TO "response_token";
ALTER TABLE conv_turns ADD COLUMN request_token int DEFAULT 0;
-- +goose StatementEnd
