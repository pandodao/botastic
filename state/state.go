package state

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/pandodao/botastic/api"
	"github.com/pandodao/botastic/internal/llms"
	llmapi "github.com/pandodao/botastic/internal/llms/api"
	"github.com/pandodao/botastic/models"
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
	turnsChan chan *models.Turn
	sh        *storage.Handler
	llms      *llms.Handler

	conversationsLock sync.Mutex
	conversations     map[uuid.UUID]*conversation
}

func New(sh *storage.Handler) *Handler {
	return &Handler{
		sh:        sh,
		turnsChan: make(chan *models.Turn),
	}
}

func (h *Handler) Start() error {
	return nil
}

func (h *Handler) handleTurnsWorker(ctx context.Context) {
	turn := <-h.turnsChan

	var c *conversation
	result, err := func() (*llmapi.ChatResponse, error) {
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

	var updateErr error
	if err != nil {
		var target *api.TurnError
		if !errors.As(err, &target) {
			target = api.NewTurnError(api.TurnErrorCodeInternalServer, err.Error())
		}

		turn.Status = api.TurnStatusFailed
		updateErr = h.sh.UpdateTurnToFailed(ctx, turn.ID, target, models.MiddlewareResults{})
	} else {
		turn.Response = result.Response
		turn.Status = api.TurnStatusSuccess
		turn.PromptTokens = result.PromptTokens
		turn.CompletionTokens = result.CompletionTokens
		turn.TotalTokens = result.TotalTokens
		turn.MiddlewareResults = models.MiddlewareResults{}
		updateErr = h.sh.UpdateTurnToSuccess(ctx, turn.ID, turn.Response, turn.PromptTokens, turn.CompletionTokens, turn.TotalTokens, turn.MiddlewareResults)
	}

	if updateErr != nil {
		// panic if we can't update the turn
		log.Panicf("failed to update turn: %v, process err: %v", updateErr, err)
	}

	if turn.Status == api.TurnStatusSuccess {
		conv.Lock()
	}
}

func (h *Handler) getOrloadConversation(ctx context.Context, convID uuid.UUID) (*conversation, error) {
	conv, err := h.sh.GetConv(ctx, convID)
	if err != nil {
		return nil, err
	}
	if conv == nil {
		return nil, api.NewTurnError(api.TurnErrorCodeConvNotFound)
	}

	h.conversationsLock.Lock()
	c, ok := h.conversations[convID]
	h.conversationsLock.Unlock()

	if !ok {
		turns, err := h.sh.GetTurns(ctx, convID, api.TurnStatusSuccess, 100)
		if err != nil {
			return nil, err
		}

		c = &conversation{
			history: turns,
		}

		h.conversationsLock.Lock()
		h.conversations[convID] = c
		h.conversationsLock.Unlock()
	}

	c.Lock()
	defer c.Unlock()
	c.conv = conv

	return c, nil
}
