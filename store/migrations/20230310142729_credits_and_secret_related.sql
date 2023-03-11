-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE users ADD COLUMN "credits" numeric(64,8);
ALTER TABLE conv_turns ADD COLUMN "user_id" bigint;

ALTER TABLE conv_turns DROP COLUMN IF EXISTS "request_token";
ALTER TABLE conv_turns RENAME COLUMN "response_token" TO "total_tokens";

CREATE TABLE "bots" (
  "id" BIGSERIAL PRIMARY KEY,
  "user_id" bigint,
  "name" varchar(128),
  "model" varchar(128) default '',
  "temperature" float default 1.0,
  "max_turn_count" int default 8,
  "context_turn_count" int default 8,
  "prompt" text default '',
  "middleware_json" jsonb default '{}',
  "public" boolean default 'f',

  "created_at" timestamptz,
  "updated_at" timestamptz,
  "deleted_at" timestamptz
);
CREATE INDEX idx_bots_user_id ON "bots" USING BTREE("user_id");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE users DROP COLUMN IF EXISTS "credits";
ALTER TABLE conv_turns DROP COLUMN IF EXISTS "user_id";

ALTER TABLE conv_turns RENAME COLUMN "total_tokens" TO "response_token";
ALTER TABLE conv_turns ADD COLUMN request_token int DEFAULT 0;

DROP TABLE IF EXISTS "bots";
-- +goose StatementEnd
