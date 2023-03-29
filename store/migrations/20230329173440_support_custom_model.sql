-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE models ADD COLUMN "custom_config" jsonb DEFAULT '{}';
ALTER TABLE models ADD COLUMN "function" varchar(64) DEFAULT '';

UPDATE models SET function = 'chat' WHERE provider = 'openai' AND provider_model IN ('gpt-4','gpt-4-32k','gpt-3.5-turbo','text-davinci-003');
UPDATE models SET function = 'embedding' WHERE provider = 'openai' AND provider_model = 'text-embedding-ada-002';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE models DROP COLUMN IF EXISTS "custom_config";
ALTER TABLE models DROP COLUMN IF EXISTS "function";
-- +goose StatementEnd
