-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE INDEX idx_orders_status ON "orders" USING BTREE("status");
CREATE INDEX idx_conv_turns_status ON "conv_turns" USING BTREE("status");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_conv_turns_status;
-- +goose StatementEnd
