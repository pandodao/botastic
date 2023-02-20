-- +goose Up
-- +goose StatementBegin
INSERT INTO "apps" ("app_id", "app_secret_encrypted", "created_at", "updated_at", "deleted_at") VALUES ( 'a4a25815-029b-42b0-be3f-5d72184a2d09', '3b3ceb5de508ef260e1d57b72d108913b8e3c93782ca7ab3ee5a66a3582803f0431ad3ae92dc98dda7d299df6329d5e77d2b64c7fe001ba100e86ca69f003938aba9e9b74769eacd0f1b3a88f8d0ee234c78729056a8dbb417184541', '2023-02-20 16:37:51.066553+09', '2023-02-20 16:37:51.066553+09', NULL);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM "app";
-- +goose StatementEnd
