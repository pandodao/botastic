-- +goose Up
-- +goose StatementBegin
CREATE TABLE properties (
    key character varying(255) PRIMARY KEY,
    value character varying(255),
    updated_at timestamp with time zone
);

CREATE TABLE apps (
    id BIGSERIAL PRIMARY KEY,
    app_id varchar(256) NOT NULL UNIQUE,
    app_secret_encrypted varchar(256) NOT NULL,

    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz
);
CREATE INDEX idx_apps_app_id ON "apps" USING BTREE("app_id", "deleted_at");

CREATE TABLE conv_turns (
    id BIGSERIAL PRIMARY KEY,
    conversation_id uuid NOT NULL,
    bot_id bigint NOT NULL,
    app_id bigint NOT NULL,
    user_identity varchar(256) NOT NULL,
    request text NOT NULL,
    response text NOT NULL DEFAULT '',
    status int NOT NULL DEFAULT 0,
    created_at timestamptz,
    updated_at timestamptz
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS properties;
DROP TABLE IF EXISTS apps;
DROP TABLE IF EXISTS conv_turns;
-- +goose StatementEnd
