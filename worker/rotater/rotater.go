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
	"github.com/pandodao/botastic/internal/tokencal"
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
		tokencal    *tokencal.Handler
	}

	TurnRequest struct {
		TurnID      uint64
		Request     *gogpt.CompletionRequest
		ChatRequest *gogpt.ChatCompletionRequest
	}
)

func New(
	cfg Config,
	gptHandler *gpt.Handler,
	convs core.ConversationStore,
	convz core.ConversationService,
	botz core.BotService,
	tokencal *tokencal.Handler,
) *Worker {
	turnReqChan := make(chan TurnRequest)
	return &Worker{
		cfg:         cfg,
		gptHandler:  gptHandler,
		convs:       convs,
		convz:       convz,
		botz:        botz,
		turnReqChan: turnReqChan,
		tokencal:    tokencal,
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

		turnReq := TurnRequest{
			TurnID: turn.ID,
		}

		switch bot.Model {
		case "gpt-3.5-turbo", "gpt-3.5-turbo-0301":
			// chat completion
			request := gogpt.ChatCompletionRequest{
				Model:    bot.Model,
				Messages: bot.GetChatMessages(conv),
			}

			turnReq.ChatRequest = &request
		default:
			// text completion
			prompt := bot.GetPrompt(conv, turn.Request)
			request := gogpt.CompletionRequest{
				Model:       bot.Model,
				Prompt:      prompt,
				MaxTokens:   1024,
				Temperature: 1,
				Stop:        []string{"Q:"},
				User:        conv.GetKey(),
			}
			turnReq.Request = &request
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

		if err := w.convs.UpdateConvTurn(ctx, turn.ID, "", 0, core.ConvTurnStatusPending); err != nil {
			continue
		}

		w.turnReqChan <- turnReq
	}
	return nil
}

func (w *Worker) UpdateConvTurnAsError(ctx context.Context, id uint64, errMsg string) error {
	fmt.Printf("errMsg: %v\n", errMsg)
	if err := w.convs.UpdateConvTurn(ctx, id, "Something wrong happened", 0, core.ConvTurnStatusError); err != nil {
		return err
	}
	return nil
}

func (w *Worker) subworker(ctx context.Context, id int) {
	for {
		// Wait for a request to handle.
		turnReq := <-w.turnReqChan

		respText := ""
		var err error
		switch {
		case turnReq.Request != nil:
			var gptResp gogpt.CompletionResponse
			gptResp, err = w.gptHandler.CreateCompletion(ctx, *turnReq.Request)
			if err == nil {
				prefix := "A:"
				respText = strings.TrimPrefix(gptResp.Choices[0].Text, prefix)
				respText = strings.TrimSpace(respText)
			}
		case turnReq.ChatRequest != nil:
			var gptResp gogpt.ChatCompletionResponse
			gptResp, err = w.gptHandler.CreateChatCompletion(ctx, *turnReq.ChatRequest)
			if err == nil {
				respText = strings.TrimSpace(gptResp.Choices[0].Message.Content)
			}
		}

		if err != nil {
			w.UpdateConvTurnAsError(ctx, turnReq.TurnID, err.Error())
			continue
		}

		// TODO use the usage field in response
		token, err := w.tokencal.GetToken(ctx, respText)
		if err != nil {
			continue
		}

		if err := w.convs.UpdateConvTurn(ctx, turnReq.TurnID, respText, token, core.ConvTurnStatusCompleted); err != nil {
			continue
		}

		if turnReq.Request != nil {
			fmt.Printf("[%03d]prompt: %+v\n", id, turnReq.Request.Prompt)
		} else if turnReq.ChatRequest != nil {
			msgs := []string{}
			for _, m := range turnReq.ChatRequest.Messages {
				msgs = append(msgs, m.Content)
			}
			fmt.Printf("[%03d]prompt: %+v\n", id, strings.Join(msgs, "\n"))
		}
		fmt.Printf("[%03d]resp: %+v\n", id, respText)

		currentRequestsLock.Lock()
		currentRequests--
		currentRequestsLock.Unlock()
	}
}
