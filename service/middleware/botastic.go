package middleware

import (
	"context"
	"fmt"

	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/session"
	"github.com/pandodao/botastic/worker/rotater"
)

type botastic struct {
	rotater *rotater.Worker
	bots    core.BotStore
}

type botasticOptions struct {
	BotID uint64
}

func (m *botastic) Name() string {
	return core.MiddlewareBotastic
}

func (m *botastic) ValidateOptions(opts map[string]any) (any, error) {
	options := &botasticOptions{}

	if val, ok := opts["bot_id"]; ok {
		v, ok := val.(float64)
		if !ok {
			return nil, fmt.Errorf("bot_id is not a number: %v", val)
		}

		if v <= 0 || float64(int(v)) != v {
			return nil, fmt.Errorf("bot_id is not a positive integer: %v", v)
		}
		options.BotID = uint64(v)
	}

	return options, nil
}

func (m *botastic) Process(ctx context.Context, opts any, input string) (string, error) {
	options := opts.(*botasticOptions)
	app := session.AppFrom(ctx)

	// make sure bot is exist
	bot, err := m.bots.GetBot(ctx, options.BotID)
	if err != nil {
		return "", fmt.Errorf("error when getting bot by bot_id: %d", options.BotID)
	}

	if bot.UserID != app.UserID {
		return "", fmt.Errorf("bot_id not found: %d", options.BotID)
	}

	t := &core.ConvTurn{
		BotID:        bot.ID,
		AppID:        app.ID,
		UserID:       app.UserID,
		UserIdentity: "",
		Request:      input,
		Status:       core.ConvTurnStatusInit,
	}
	if err := m.rotater.ProcessConvTurn(ctx, t); err != nil {
		return "", err
	}

	return t.Response, nil
}
