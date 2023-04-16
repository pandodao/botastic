package order

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/order"
	"github.com/pandodao/botastic/store/user"
	"github.com/pandodao/lemon-checkout-go"
	"github.com/pandodao/mixpay-go"
	"github.com/shopspring/decimal"
)

func New(
	cfg Config,
	orders core.OrderStore,
	userz core.UserService,
	mixpayClient *mixpay.Client,
	lemonClient *lemon.Client,
) *service {
	return &service{
		cfg:          cfg,
		orders:       orders,
		userz:        userz,
		mixpayClient: mixpayClient,
		lemonClient:  lemonClient,
	}
}

type (
	Config struct {
		PayeeId           string
		QuoteAssetId      string
		SettlementAssetId string
		CallbackUrl       string
		ReturnTo          string
		FailedReturnTo    string
	}

	service struct {
		cfg          Config
		orders       core.OrderStore
		userz        core.UserService
		mixpayClient *mixpay.Client
		lemonClient  *lemon.Client
	}
)

func (s *service) CreateMixpayOrder(ctx context.Context, userId uint64, amount decimal.Decimal) (string, error) {
	id := uuid.Must(uuid.NewV4())
	orderId := id.String()
	traceId := uuid.NewV5(id, "trace").String()
	err := s.orders.CreateOrder(ctx, orderId, userId, core.OrderChannelMixpay, core.OrderStatusPending, s.cfg.PayeeId, amount, s.cfg.QuoteAssetId, traceId)
	if err != nil {
		return "", err
	}

	resp, err := s.mixpayClient.CreateOneTimePayment(ctx, mixpay.CreateOneTimePaymentRequest{
		PayeeId:           s.cfg.PayeeId,
		QuoteAssetId:      s.cfg.QuoteAssetId,
		QuoteAmount:       amount.String(),
		SettlementAssetId: s.cfg.SettlementAssetId,
		OrderId:           orderId,
		TraceId:           traceId,
		CallbackUrl:       s.cfg.CallbackUrl,
		ReturnTo:          s.cfg.ReturnTo,
		FailedReturnTo:    s.cfg.FailedReturnTo,
	})
	if err != nil {
		return "", err
	}

	return resp.Code, nil
}

func (s *service) CreateLemonOrder(ctx context.Context, userID uint64, storeID int64, variantID int64, amount decimal.Decimal, redirectURL string) (string, error) {
	id := uuid.Must(uuid.NewV4())
	orderID := id.String()
	traceId := uuid.NewV5(id, "trace").String()
	err := s.orders.CreateOrder(ctx, orderID, userID, core.OrderChannelLemon, core.OrderStatusPending, s.cfg.PayeeId, amount, s.cfg.QuoteAssetId, traceId)
	if err != nil {
		return "", err
	}

	var custom struct {
		UserID  string `json:"user_id"`
		OrderID string `json:"order_id"`
	}
	custom.UserID = strconv.FormatUint(userID, 10)
	custom.OrderID = orderID

	resp, err := s.lemonClient.CreateCheckoutSimple(ctx, storeID, []int64{variantID}, redirectURL, lemon.CheckoutData{
		Custom: custom,
	})
	if err != nil {
		return "", err
	}

	return resp.Attributes.URL, nil
}

func (s *service) HandleMixpayCallback(ctx context.Context, orderId string, traceId string, payeeId string) error {
	o, err := s.orders.GetOrder(ctx, orderId)
	if err != nil {
		return err
	}
	if o.Status != core.OrderStatusPending {
		return nil
	}

	resp, err := s.mixpayClient.GetPaymentResult(ctx, mixpay.GetPaymentResultRequest{
		OrderId: orderId,
		TraceId: traceId,
		PayeeId: payeeId,
	})
	if err != nil {
		return err
	}

	if resp.Status == o.UpstreamStatus {
		return nil
	}

	status := core.OrderStatusPending
	switch resp.Status {
	case "failed":
		status = core.OrderStatusFailed
	case "success":
		status = core.OrderStatusSuccess
		if resp.PayeeID != o.PayeeId {
			return fmt.Errorf("invalid payee id: expected %s, got %s", o.PayeeId, resp.PayeeID)
		}
		if resp.QuoteAssetID != o.QuoteAssetId {
			return fmt.Errorf("invalid quote asset id: expected %s, got %s", o.QuoteAssetId, resp.QuoteAssetID)
		}
		if resp.QuoteAmount != o.QuoteAmount.String() {
			return fmt.Errorf("invalid quote amount: expected %s, got %s", o.QuoteAmount.String(), resp.QuoteAmount)
		}
	}

	if err := store.Transaction(func(h *store.Handler) error {
		orders := order.New(h)
		users := user.New(h)
		userz := s.userz.ReplaceStore(users)

		if err := orders.UpdateOrder(ctx, orderId, resp.Status, status, resp.Raw); err != nil {
			return err
		}

		if status == core.OrderStatusSuccess {
			// TODO select for update
			user, err := users.GetUser(ctx, o.UserID)
			if err != nil {
				return err
			}

			// update user credit
			if err := userz.Topup(ctx, user, o.QuoteAmount); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *service) HandleLemonCallback(ctx context.Context, orderID string, userID uint64, lemonAmount decimal.Decimal, upstreamStatus string) error {
	od, err := s.orders.GetOrder(ctx, orderID)
	if err != nil || od.UserID != userID || od.Status != core.OrderStatusPending {
		return err
	}

	status := core.OrderStatusPending
	switch upstreamStatus {
	case "cancelled":
		status = core.OrderStatusCanceled
	case "paid":
		status = core.OrderStatusSuccess
		diff := od.QuoteAmount.Sub(lemonAmount)
		if diff.Abs().GreaterThan(decimal.NewFromFloat(0.01)) {
			return fmt.Errorf("invalid amount: expected %s, got %s", od.QuoteAmount.StringFixed(8), lemonAmount.StringFixed(8))
		}
	}

	if err := store.Transaction(func(h *store.Handler) error {
		orders := order.New(h)
		users := user.New(h)
		userz := s.userz.ReplaceStore(users)

		if status == core.OrderStatusSuccess {
			if err := orders.UpdateOrder(ctx, orderID, upstreamStatus, core.OrderStatusSuccess, ""); err != nil {
				return err
			}

			// TODO select for update
			user, err := users.GetUser(ctx, od.UserID)
			if err != nil {
				return err
			}

			// update user credit
			if err := userz.Topup(ctx, user, od.QuoteAmount); err != nil {
				return err
			}
		} else if status == core.OrderStatusCanceled {
			if err := orders.UpdateOrder(ctx, orderID, upstreamStatus, core.OrderStatusCanceled, ""); err != nil {
				return err
			}

		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
