package rotater

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fox-one/pkg/logger"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/internal/gpt"
	gogpt "github.com/sashabaranov/go-gpt3"
)

type (
	Config struct {
	}

	Worker struct {
		cfg        Config
		gptHandler *gpt.Handler
		convs      core.ConversationStore
		convz      core.ConversationService
		botz       core.BotService
	}
)

func New(
	cfg Config,
	gptHandler *gpt.Handler,
	convs core.ConversationStore,
	convz core.ConversationService,
	botz core.BotService,
) *Worker {

	return &Worker{
		cfg:        cfg,
		gptHandler: gptHandler,
		convs:      convs,
		convz:      convz,
		botz:       botz,
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
		bot, err := w.botz.GetBot(ctx, turn.BotID)
		if err != nil {
			w.UpdateConvTurnAsError(ctx, turn.ID, err.Error())
			continue
		}

		conv, err := w.convz.GetConversation(ctx, turn.ConversationID)
		if err != nil {
			w.UpdateConvTurnAsError(ctx, turn.ID, err.Error())
			continue
		}

		prompt := bot.GetPrompt(conv, turn.Request)

		// @TODO send to OpenAPI concurrently
		request := gogpt.CompletionRequest{
			Model:       bot.Model,
			Prompt:      prompt,
			MaxTokens:   1024,
			Temperature: 1,
			Stop:        []string{"Q:"},
			User:        conv.GetKey(),
		}
		gptResp, err := w.gptHandler.CreateCompletion(ctx, request)
		if err != nil {
			w.UpdateConvTurnAsError(ctx, turn.ID, err.Error())
			continue
		}

		prefix := "A:"
		respText := strings.TrimPrefix(gptResp.Choices[0].Text, prefix)
		respText = strings.TrimSpace(respText)
		if err := w.convs.UpdateConvTurn(ctx, turn.ID, respText, core.ConvTurnStatusCompleted); err != nil {
			continue
		}
		fmt.Printf("prompt: %v\n", prompt)
		fmt.Printf("resp: %+v\n", respText)
	}
	return nil
}

func (w *Worker) UpdateConvTurnAsError(ctx context.Context, id uint64, errMsg string) error {
	fmt.Printf("errMsg: %v\n", errMsg)
	if err := w.convs.UpdateConvTurn(ctx, id, "Something wrong happened", core.ConvTurnStatusError); err != nil {
		return err
	}
	return nil
}
