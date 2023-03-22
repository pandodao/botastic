package ordersyncer

import (
	"context"
	"time"

	"github.com/fox-one/pkg/logger"
	"github.com/pandodao/botastic/core"
)

type Config struct {
	Interval       time.Duration
	CheckInterval  time.Duration
	CancelInterval time.Duration
}

type Worker struct {
	cfg    Config
	orders core.OrderStore
	orderz core.OrderService
}

func New(
	cfg Config,
	orders core.OrderStore,
	orderz core.OrderService,
) *Worker {
	return &Worker{
		cfg:    cfg,
		orderz: orderz,
		orders: orders,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "ordersyncer")
	ctx = logger.WithContext(ctx, log)

	for {
		if err := w.run(ctx); err != nil {
			log.WithError(err).Errorln("ordersyncer run failed")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(w.cfg.Interval):
		}
	}
}

func (w *Worker) run(ctx context.Context) error {
	os, err := w.orders.GetOrdersByStatus(ctx, core.OrderStatusPending)
	if err != nil {
		return err
	}
	log := logger.FromContext(ctx)
	for _, o := range os {
		if o.Channel != core.OrderChannelMixpay {
			continue
		}

		if func() error {
			now := time.Now()
			if o.CreatedAt.Add(w.cfg.CancelInterval).Before(now) {
				return w.orders.UpdateOrderStatus(ctx, o.ID, core.OrderStatusCanceled)
			}

			if o.CreatedAt.Add(w.cfg.CheckInterval).Before(now) {
				return w.orderz.HandleMixpayCallback(ctx, o.ID, o.TraceID, o.PayeeId)
			}

			return nil
		}(); err != nil {
			log.WithError(err).Errorln("order sync failed, id:", o.ID)
		}
	}
	return nil
}
