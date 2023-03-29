-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
UPDATE bots SET model = CONCAT('openai:', model);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
UPDATE bots SET model = SUBSTR(model, 8);
-- +goose StatementEnd
