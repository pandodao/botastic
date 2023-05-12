package state

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pandodao/botastic/api"
	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/internal/llms"
	llmapi "github.com/pandodao/botastic/internal/llms/api"
	"github.com/pandodao/botastic/models"
	"github.com/pandodao/botastic/pkg/chanhub"
	"github.com/pandodao/botastic/storage"
)

type conversation struct {
	sync.Mutex
	conv    *models.Conv
	history []*models.Turn
}

func (c *conversation) historyText() []string {
	if len(c.history) == 0 {
		return []string{}
	}

	text := make([]string, 0, len(c.history)*2)
	for _, t := range c.history {
		text = append(text, t.Request)
		text = append(text, t.Response)
	}
	return text
}

type Handler struct {
	cfg       config.StateConfig
	turnsChan chan *models.Turn
	sh        *storage.Handler
	llms      *llms.Handler
	hub       *chanhub.Hub

	conversationsLock sync.Mutex
	conversations     map[uuid.UUID]*conversation
}

func New(cfg config.StateConfig, sh *storage.Handler, llms *llms.Handler, hub *chanhub.Hub) *Handler {
	return &Handler{
		cfg:           cfg,
		sh:            sh,
		llms:          llms,
		turnsChan:     make(chan *models.Turn),
		conversations: make(map[uuid.UUID]*conversation),
		hub:           hub,
	}
}

func (h *Handler) Start(ctx context.Context) error {
	turns, err := h.sh.GetTurnsByStatus(ctx, []api.TurnStatus{api.TurnStatusInit, api.TurnStatusProcessing})
	if err != nil {
		return err
	}

	initTurns := []*models.Turn{}
	for _, turn := range turns {
		if turn.Status == api.TurnStatusInit {
			initTurns = append(initTurns, turn)
		} else {
			if err := h.sh.UpdateTurnToFailed(ctx, turn.ID, api.NewTurnError(api.TurnErrorCodeInternalServer, "turn processing terminated unexpectedly"), nil); err != nil {
				return err
			}
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(h.cfg.WorkerCount)

	for i := 0; i < h.cfg.WorkerCount; i++ {
		go func() {
			defer wg.Done()
			h.handleTurnsWorker(ctx)
		}()
	}

	for _, turn := range initTurns {
		h.turnsChan <- turn
	}

	wg.Wait()
	return nil
}

func (h *Handler) GetTurnsChan() chan<- *models.Turn {
	return h.turnsChan
}

func (h *Handler) handleTurnsWorker(ctx context.Context) {
	var turn *models.Turn
	select {
	case <-ctx.Done():
		return
	case turn = <-h.turnsChan:
	}

	var c *conversation
	result, err := func() (*llmapi.ChatResponse, error) {
		if err := h.sh.UpdateTurnToProcessing(ctx, turn.ID); err != nil {
			return nil, err
		}

		var err error
		c, err = h.getOrloadConversation(ctx, turn.ConvID)
		if err != nil {
			return nil, err
		}

		c.Lock()
		defer c.Lock()

		bot, err := h.sh.GetBot(ctx, turn.BotID)
		if err != nil {
			return nil, err
		}
		if bot == nil {
			return nil, api.NewTurnError(api.TurnErrorCodeBotNotFound)
		}

		cm, ok := h.llms.GetChatModel(bot.ChatModel)
		if !ok {
			return nil, api.NewTurnError(api.TurnErrorCodeChatModelNotFound)
		}

		result, err := cm.Chat(ctx, llmapi.ChatRequest{
			Temperature:    bot.Temperature,
			Prompt:         bot.Prompt,
			BoundaryPrompt: bot.BoundaryPrompt,
			History:        c.historyText(),
			Request:        turn.Request,
		})
		if err != nil {
			return nil, api.NewTurnError(api.TurnErrorCodeChatModelCallError, err.Error())
		}

		return result, nil
	}()

	var updateFunc func() error
	if err != nil {
		var target *api.TurnError
		if !errors.As(err, &target) {
			target = api.NewTurnError(api.TurnErrorCodeInternalServer, err.Error())
		}

		turn.Status = api.TurnStatusFailed
		updateFunc = func() error {
			return h.sh.UpdateTurnToFailed(ctx, turn.ID, target, models.MiddlewareResults{})
		}
	} else {
		turn.Response = result.Response
		turn.Status = api.TurnStatusSuccess
		turn.PromptTokens = result.PromptTokens
		turn.CompletionTokens = result.CompletionTokens
		turn.TotalTokens = result.TotalTokens
		turn.MiddlewareResults = models.MiddlewareResults{}
		updateFunc = func() error {
			return h.sh.UpdateTurnToSuccess(ctx, turn.ID, turn.Response, turn.PromptTokens, turn.CompletionTokens, turn.TotalTokens, turn.MiddlewareResults)
		}
	}

	for {
		updateErr := updateFunc()
		if err == nil {
			break
		}

		log.Printf("failed to update turn: %v, process err: %v\n", updateErr, err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Second):
		}
	}

	h.hub.Broadcast(turn.ID, struct{}{})
}

func (h *Handler) getOrloadConversation(ctx context.Context, convID uuid.UUID) (*conversation, error) {
	conv, err := h.sh.GetConv(ctx, convID)
	if err != nil {
		return nil, err
	}
	if conv == nil {
		return nil, api.NewTurnError(api.TurnErrorCodeConvNotFound)
	}

	c, err := func() (*conversation, error) {
		h.conversationsLock.Lock()
		defer h.conversationsLock.Unlock()

		c, ok := h.conversations[convID]
		if !ok {
			turns, err := h.sh.GetTurns(ctx, convID, api.TurnStatusSuccess, 100)
			if err != nil {
				return nil, err
			}

			c = &conversation{
				history: turns,
			}

			h.conversations[convID] = c
		}
		return c, nil
	}()
	if err != nil {
		return nil, err
	}

	c.Lock()
	defer c.Unlock()
	c.conv = conv

	return c, nil
}
