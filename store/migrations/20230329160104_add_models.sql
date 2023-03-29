-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE "models" (
  "id" BIGSERIAL PRIMARY KEY,
  "provider" varchar(64) DEFAULT '',
  "provider_model" varchar(64) DEFAULT '',
  "max_token" int default 0,
  "prompt_price_usd" numeric(64,8) DEFAULT 0,
  "completion_price_usd" numeric(64,8) DEFAULT 0,
  "price_usd" numeric(64,8) DEFAULT 0,
  "created_at" timestamptz,
  "deleted_at" timestamptz
);

CREATE INDEX idx_models_deleted_at ON "models" USING BTREE("deleted_at");
CREATE UNIQUE INDEX idx_models_provider_provider_model ON models("provider", "provider_model") WHERE "deleted_at" IS NULL;

INSERT INTO "models"("provider","provider_model","max_token","prompt_price_usd","completion_price_usd","price_usd","created_at","deleted_at")
VALUES
(E'openai',E'gpt-4',8192,0.00003,0.00006,0,NOW(),NULL),
(E'openai',E'gpt-4-32k',32768,0.00006,0.00012,0,NOW(),NULL),
(E'openai',E'gpt-3.5-turbo',4096,0,0,0.000002,NOW(),NULL),
(E'openai',E'text-davinci-003',4097,0,0,0.00002,NOW(),NULL),
(E'openai',E'text-embedding-ada-002',8191,0,0,0.0000004,NOW(),NULL);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS "models";
-- +goose StatementEnd
