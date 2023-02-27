-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE conv_turns ADD COLUMN request_token int DEFAULT 0;
ALTER TABLE conv_turns ADD COLUMN response_token int DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE conv_turns DROP COLUMN IF EXISTS request_token;
ALTER TABLE conv_turns DROP COLUMN IF EXISTS response_token;
-- +goose StatementEnd
