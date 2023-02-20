package rotater

import (
	"context"
	"time"

	"github.com/fox-one/pkg/logger"
	"github.com/pandodao/botastic/core"
)

type (
	Config struct {
	}

	Worker struct {
		cfg   Config
		convs core.ConversationStore
	}
)

func New(
	cfg Config,
	convs core.ConversationStore,
) *Worker {

	return &Worker{
		cfg:   cfg,
		convs: convs,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "rotater")
	ctx = logger.WithContext(ctx, log)
	log.Println("start rotater worker")

	dur := time.Millisecond
	var circle int64
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(dur):
			if err := w.run(ctx, circle); err == nil {
				dur = time.Second
				circle += 1
			} else {
				dur = 10 * time.Second
				circle += 10
			}
		}
	}
}

func (w *Worker) run(ctx context.Context, circle int64) error {
	turns, err := w.convs.GetConvTurnsByStatus(ctx, core.ConvTurnStatusInit)
	if err != nil {
		return err
	}

	for _, turn := range turns {
		// @TODO send to chatGPT and get response
		if err := w.convs.UpdateConvTurn(ctx, turn.ID, "Pong!", core.ConvTurnStatusCompleted); err != nil {
			continue
		}
	}
	return nil
}
