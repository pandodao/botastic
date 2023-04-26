package worker

import (
	"context"

	"github.com/pandodao/botastic/core"
)

type Worker interface {
	Run(ctx context.Context) error
}

type Rotater interface {
	ProcessConvTurn(ctx context.Context, turn *core.ConvTurn) error
}
