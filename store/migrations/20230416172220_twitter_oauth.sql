-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE "users" ADD COLUMN "email" varchar(64);
ALTER TABLE "users" ADD COLUMN "twitter_id" varchar(32);
ALTER TABLE "users" ADD COLUMN "twitter_screen_name" varchar(128);
CREATE INDEX idx_users_twitter_id ON "users" USING BTREE("twitter_id", "deleted_at");
CREATE INDEX idx_users_email ON "users" USING BTREE("email", "deleted_at");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE "users" DROP COLUMN "email";
ALTER TABLE "users" DROP COLUMN "twitter_id";
ALTER TABLE "users" DROP COLUMN "twitter_screen_name";
DROP INDEX IF EXISTS idx_users_twitter_id;
DROP INDEX IF EXISTS idx_users_email
-- +goose StatementEnd
