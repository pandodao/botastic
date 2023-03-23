package order

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/internal/mixpay"
	"github.com/pandodao/botastic/store"
	"github.com/pandodao/botastic/store/order"
	"github.com/pandodao/botastic/store/user"
	"github.com/shopspring/decimal"
)

func New(
	cfg Config,
	orders core.OrderStore,
	userz core.UserService,
	mixpayClient *mixpay.Client,
) *service {
	return &service{
		cfg:          cfg,
		orders:       orders,
		userz:        userz,
		mixpayClient: mixpayClient,
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
