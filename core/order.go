package core

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type (
	OrderStatus  string
	OrderChannel string
)

const (
	OrderStatusPending  OrderStatus = "PENDING"
	OrderStatusSuccess  OrderStatus = "SUCCESS"
	OrderStatusFailed   OrderStatus = "FAILED"
	OrderStatusCanceled OrderStatus = "CANCELED"

	OrderChannelMixpay OrderChannel = "Mixpay"
	OrderChannelLemon  OrderChannel = "Lemon"
)

type Order struct {
	ID             string
	UserID         uint64
	Channel        OrderChannel
	Status         OrderStatus
	PayeeId        string
	QuoteAmount    decimal.Decimal
	QuoteAssetId   string
	TraceID        string
	UpstreamStatus string
	Raw            string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type OrderStore interface {

	// INSERT INTO @@table
	// 	("id", "user_id", "channel", "status", "payee_id", "quote_amount", "quote_asset_id", "trace_id", "created_at", "updated_at")
	// VALUES
	// 	(@id, @userId, @channel, @status, @payeeId, @quoteAmount, @quoteAssetId, @traceId, NOW(), NOW())
	CreateOrder(ctx context.Context, id string, userId uint64, channel OrderChannel, status OrderStatus, payeeId string,
		quoteAmount decimal.Decimal, quoteAssetId, traceId string) error

	// UPDATE @@table
	// 	{{set}}
	// 		"upstream_status"=@upstreamStatus,
	// 		"status"=@status,
	// 		"raw"=@raw,
	// 		"updated_at"=NOW()
	// 	{{end}}
	// WHERE
	// 	"id"=@id
	UpdateOrder(ctx context.Context, id string, upstreamStatus string, status OrderStatus, raw string) error

	// SELECT *
	// FROM @@table WHERE
	// 	"id"=@id
	// LIMIT 1
	GetOrder(ctx context.Context, id string) (*Order, error)

	// SELECT *
	// FROM @@table WHERE
	// 	"status"=@status
	GetOrdersByStatus(ctx context.Context, status OrderStatus) ([]*Order, error)

	// UPDATE @@table
	// 	{{set}}
	// 		"status"=@status,
	// 		"updated_at"=NOW()
	// 	{{end}}
	// WHERE
	// 	"id"=@id
	UpdateOrderStatus(ctx context.Context, id string, status OrderStatus) error
}

type OrderService interface {
	CreateMixpayOrder(ctx context.Context, userId uint64, amount decimal.Decimal) (string, error)
	CreateLemonOrder(ctx context.Context, userID uint64, storeID int64, variantID int64, amount decimal.Decimal, redirectURL string) (string, error)
	HandleMixpayCallback(ctx context.Context, orderId string, traceId string, payeeId string) error
	HandleLemonCallback(ctx context.Context, orderID string, userID uint64, lemonAmount decimal.Decimal, upstreamStatus string) error
}
