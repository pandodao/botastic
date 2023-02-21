package rotater

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fox-one/pkg/logger"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/internal/gpt"
	gogpt "github.com/sashabaranov/go-gpt3"
)

const (
	MAX_SUBWORKERS          = 32
	MAX_REQUESTS_PER_MINUTE = 3000
)

var (
	currentRequests     int
	currentRequestsLock sync.Mutex
)

type (
	Config struct {
	}

	Worker struct {
		cfg         Config
		gptHandler  *gpt.Handler
		convs       core.ConversationStore
		convz       core.ConversationService
		botz        core.BotService
		turnReqChan chan TurnRequest
	}

	TurnRequest struct {
		TurnID  uint64
		Request gogpt.CompletionRequest
	}
)

func New(
	cfg Config,
	gptHandler *gpt.Handler,
	convs core.ConversationStore,
	convz core.ConversationService,
	botz core.BotService,
) *Worker {
	turnReqChan := make(chan TurnRequest)
	return &Worker{
		cfg:         cfg,
		gptHandler:  gptHandler,
		convs:       convs,
		convz:       convz,
		botz:        botz,
		turnReqChan: turnReqChan,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	log := logger.FromContext(ctx).WithField("worker", "rotater")
	ctx = logger.WithContext(ctx, log)
	log.Println("start rotater subworkers")
	for i := 0; i < MAX_SUBWORKERS; i++ {
		go w.subworker(ctx, i)
	}

	dur := time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(dur):
			if err := w.run(ctx); err == nil {
				dur = time.Second
			} else {
				dur = 10 * time.Second
			}
		}
	}
}

func (w *Worker) run(ctx context.Context) error {
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

		request := gogpt.CompletionRequest{
			Model:       bot.Model,
			Prompt:      prompt,
			MaxTokens:   1024,
			Temperature: 1,
			Stop:        []string{"Q:"},
			User:        conv.GetKey(),
		}

		for {
			currentRequestsLock.Lock()
			if currentRequests < MAX_REQUESTS_PER_MINUTE {
				currentRequests++
				currentRequestsLock.Unlock()
				break
			}
			currentRequestsLock.Unlock()
			time.Sleep(time.Second)
		}

		if err := w.convs.UpdateConvTurn(ctx, turn.ID, "", core.ConvTurnStatusPending); err != nil {
			continue
		}

		turnReq := TurnRequest{
			TurnID:  turn.ID,
			Request: request,
		}

		w.turnReqChan <- turnReq
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

func (w *Worker) subworker(ctx context.Context, id int) {
	for {
		// Wait for a request to handle.
		turnReq := <-w.turnReqChan

		gptResp, err := w.gptHandler.CreateCompletion(ctx, turnReq.Request)
		if err != nil {
			w.UpdateConvTurnAsError(ctx, turnReq.TurnID, err.Error())
			continue
		}

		prefix := "A:"
		respText := strings.TrimPrefix(gptResp.Choices[0].Text, prefix)
		respText = strings.TrimSpace(respText)
		if err := w.convs.UpdateConvTurn(ctx, turnReq.TurnID, respText, core.ConvTurnStatusCompleted); err != nil {
			continue
		}

		fmt.Printf("[%03d]prompt: %v\n", id, turnReq.Request.Prompt)
		fmt.Printf("[%03d]resp: %+v\n", id, respText)

		currentRequestsLock.Lock()
		currentRequests--
		currentRequestsLock.Unlock()
	}
}
