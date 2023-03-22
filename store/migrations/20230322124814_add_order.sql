-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE "orders" (
  "id" uuid PRIMARY KEY,
  "user_id" bigint,
  "channel" varchar(16) DEFAULT '',
  "status" varchar(16) DEFAULT '',
  "payee_id" text DEFAULT '',
  "quote_amount" numeric(64,8) DEFAULT 0,
  "quote_asset_id" text DEFAULT '',
  "trace_id" text DEFAULT '',
  "upstream_status" varchar(16) DEFAULT '',
  "raw" text DEFAULT '',
  "created_at" timestamptz,
  "updated_at" timestamptz
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS "orders";
-- +goose StatementEnd
