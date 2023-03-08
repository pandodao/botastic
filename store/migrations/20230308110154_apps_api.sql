-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE apps ADD COLUMN "user_id" bigint;
ALTER TABLE apps ADD COLUMN "name" varchar(128);
CREATE INDEX idx_apps_user_id ON "apps" USING BTREE("app_id", "user_id", "deleted_at");

CREATE TABLE "users" (
  "id" BIGSERIAL PRIMARY KEY,
  "mixin_user_id" uuid,
  "mixin_identity_number" varchar(16),
  "avatar_url" varchar(1024) default '',
  "full_name" varchar(128) NOT NULL,
  "lang" varchar(16) default 'en',
  "mvm_public_key" varchar(255) DEFAULT '',

  "created_at" timestamptz,
  "updated_at" timestamptz,
  "deleted_at" timestamptz
);
CREATE INDEX idx_users_mixin_user_id ON "users" USING BTREE("mixin_user_id");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE apps DROP COLUMN IF EXISTS "user_id";
ALTER TABLE apps DROP COLUMN IF EXISTS "name";
DROP INDEX IF EXISTS idx_apps_user_id ;
DROP TABLE IF EXISTS "users";
-- +goose StatementEnd
