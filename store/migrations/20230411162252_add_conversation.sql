-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE "conversations" (
  "id" uuid PRIMARY KEY,
  "lang" varchar(32) DEFAULT 'en',
  "user_identity" varchar(128) DEFAULT '',
  "app_id" integer DEFAULT 0,
  "bot_id" integer DEFAULT 0,
  "created_at" timestamptz,
  "updated_at" timestamptz,
  "deleted_at" timestamptz
);

CREATE INDEX idx_conversations_deleted_at ON "conversations" USING BTREE("deleted_at");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS "conversations";
-- +goose StatementEnd
