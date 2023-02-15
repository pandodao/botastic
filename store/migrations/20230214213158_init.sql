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
    app_secret varchar(256) NOT NULL,

    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz
);
CREATE INDEX idx_apps_app_id ON "apps" USING BTREE("app_id", "deleted_at");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE properties;
DROP TABLE apps;
-- +goose StatementEnd
