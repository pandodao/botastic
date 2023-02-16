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

CREATE TABLE indexes (
    id BIGSERIAL PRIMARY KEY,
    app_id BIGSERIAL NOT NULL,
    data varchar(1024) NOT NULL,
    vectors numeric[] NOT NULL,
    object_id varchar(256) NOT NULL,
    index_name varchar(256) NOT NULL,
    category varchar(256) NOT NULL,
    properties varchar(256) NOT NULL,
    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz
);
CREATE INDEX idx_indexes_deleted_at ON "indexes" USING BTREE("app_id", "deleted_at");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS properties;
DROP TABLE IF EXISTS apps;
DROP TABLE IF EXISTS indexes;
-- +goose StatementEnd
