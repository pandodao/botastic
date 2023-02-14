-- +goose Up
-- +goose StatementBegin
CREATE TABLE properties (
    key character varying(255) PRIMARY KEY,
    value character varying(255),
    updated_at timestamp with time zone
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE properties;
-- +goose StatementEnd
