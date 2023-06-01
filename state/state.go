package state

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pandodao/botastic/api"
	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/models"
	"github.com/pandodao/botastic/pkg/chanhub"
	"github.com/pandodao/botastic/pkg/llms"
	llmapi "github.com/pandodao/botastic/pkg/llms/api"
	"github.com/pandodao/botastic/storage"
	"go.uber.org/zap"
)

type MiddlewareHandler interface {
	Process(ctx context.Context, mc api.MiddlewareConfig, turn *models.Turn) ([]*api.MiddlewareResult, bool)
}

type Handler struct {
	logger            *zap.Logger
	cfg               config.StateConfig
	turnsChan         chan *models.Turn
	sh                *storage.Handler
	llms              *llms.Handler
	hub               *chanhub.Hub
	middlewareHandler MiddlewareHandler

	tc                *templateCache
	conversationsLock sync.Mutex
	conversations     map[uuid.UUID]*conversation
}

func New(cfg config.StateConfig, logger *zap.Logger, sh *storage.Handler,
	llms *llms.Handler, hub *chanhub.Hub, middlewareHandler MiddlewareHandler) *Handler {
	return &Handler{
		logger:            logger.Named("state"),
		cfg:               cfg,
		sh:                sh,
		llms:              llms,
		turnsChan:         make(chan *models.Turn),
		conversations:     make(map[uuid.UUID]*conversation),
		hub:               hub,
		middlewareHandler: middlewareHandler,
		tc:                newTemplateCache(),
	}
}

func (h *Handler) Start(ctx context.Context) error {
	turns, err := h.sh.GetTurnsByStatus(ctx, []api.TurnStatus{api.TurnStatusInit, api.TurnStatusProcessing})
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(h.cfg.WorkerCount)

	for i := 0; i < h.cfg.WorkerCount; i++ {
		go func() {
			defer wg.Done()
			h.handleTurnsWorker(ctx)
		}()
	}

	for _, turn := range turns {
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

	h.logger.Info("handling turn", zap.Uint("turn_id", turn.ID))
	var (
		middlewareResults []*api.MiddlewareResult
		c                 *conversation
	)
	result, err := func() (*llmapi.ChatResponse, error) {
		if err := h.sh.UpdateTurnToProcessing(ctx, turn.ID); err != nil {
			return nil, err
		}
		h.logger.Debug("turn updated to processing", zap.Uint("turn_id", turn.ID))

		var err error
		c, err = h.getOrloadConversation(ctx, turn.ConvID)
		if err != nil {
			return nil, err
		}

		c.Lock()
		defer c.Unlock()

		bot, err := h.sh.GetBot(ctx, turn.BotID)
		if err != nil {
			return nil, err
		}
		if bot == nil {
			return nil, models.NewTurnError(api.TurnErrorCodeBotNotFound)
		}

		if bot.Middlewares != nil {
			var ok bool
			middlewareResults, ok = h.middlewareHandler.Process(ctx, api.MiddlewareConfig(*bot.Middlewares), turn)
			if !ok {
				return nil, models.NewTurnError(api.TurnErrorCodeMiddlewareError)
			}

			data := map[string]any{}
			for _, r := range middlewareResults {
				data[r.RenderName] = r.Result
			}

			if err := h.renderBotPrompts(bot, data); err != nil {
				return nil, models.NewTurnError(api.TurnErrorCodeRenderPromptError, err.Error())
			}
		}

		cm, ok := h.llms.GetChatModel(bot.ChatModel)
		if !ok {
			return nil, models.NewTurnError(api.TurnErrorCodeChatModelNotFound)
		}

		h.logger.Debug("chat model found", zap.String("chat_model", bot.ChatModel), zap.Uint("turn_id", turn.ID), zap.String("conv_id", turn.ConvID.String()))
		if bot.TimeoutSeconds > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Duration(bot.TimeoutSeconds)*time.Second)
			defer cancel()
		}
		result, err := cm.Chat(ctx, llmapi.ChatRequest{
			Temperature:    bot.Temperature,
			Prompt:         bot.Prompt,
			BoundaryPrompt: bot.BoundaryPrompt,
			History:        c.historyText(),
			Request:        turn.Request,
		})
		if err != nil {
			h.logger.Error("chat model error", zap.Error(err), zap.Uint("turn_id", turn.ID))
			code := api.TurnErrorCodeChatModelCallError
			if errors.Is(err, context.DeadlineExceeded) {
				code = api.TurnErrorCodeChatModelCallTimeout
			}
			return nil, models.NewTurnError(code, err.Error())
		}

		h.logger.Info("chat model response",
			zap.Uint("turn_id", turn.ID),
			zap.String("chat_model", bot.ChatModel),
			zap.Int("total_tokens", result.Usage.TotalTokens),
		)
		return result, nil
	}()

	var updateFunc func() error
	if err != nil {
		var target *models.TurnError
		if !errors.As(err, &target) {
			target = models.NewTurnError(api.TurnErrorCodeInternalServer, err.Error())
		}

		turn.Status = api.TurnStatusFailed
		updateFunc = func() error {
			return h.sh.UpdateTurnToFailed(ctx, turn.ID, target, middlewareResults)
		}
	} else {
		turn.Response = result.Response
		turn.Status = api.TurnStatusSuccess
		turn.PromptTokens = result.Usage.PromptTokens
		turn.CompletionTokens = result.Usage.CompletionTokens
		turn.TotalTokens = result.Usage.TotalTokens
		turn.MiddlewareResults = middlewareResults
		updateFunc = func() error {
			return h.sh.UpdateTurnToSuccess(ctx, turn.ID, turn.Response, turn.PromptTokens, turn.CompletionTokens, turn.TotalTokens, turn.MiddlewareResults)
		}
	}

	for {
		updateErr := updateFunc()
		if updateErr == nil {
			break
		}

		h.logger.Error("failed to update turn", zap.Error(updateErr), zap.Uint("turn_id", turn.ID))
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}

	if turn.Status == api.TurnStatusSuccess {
		c.appendTurn(turn)
	}

	h.logger.Info("turn processed", zap.Uint("turn_id", turn.ID), zap.String("status", turn.Status.String()))
	h.hub.Broadcast(turn.ID, struct{}{})
}

func (h *Handler) getOrloadConversation(ctx context.Context, convID uuid.UUID) (*conversation, error) {
	conv, err := h.sh.GetConv(ctx, convID)
	if err != nil {
		return nil, err
	}
	if conv == nil {
		return nil, models.NewTurnError(api.TurnErrorCodeConvNotFound)
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
				history: make([]*models.Turn, 0, len(turns)),
			}
			for i := len(turns) - 1; i >= 0; i-- {
				c.history = append(c.history, turns[i])
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

func (h *Handler) renderBotPrompts(b *models.Bot, data map[string]any) error {
	f := func(k, v string) (string, error) {
		if v == "" {
			return "", nil
		}
		t, err := h.tc.getTemplate(k, v)
		if err != nil {
			return "", err
		}
		var buf bytes.Buffer
		if err := t.Execute(&buf, data); err != nil {
			return "", err
		}

		return buf.String(), nil
	}

	id := strconv.Itoa(int(b.ID))
	newPrompt, err := f(id+"_prompt", b.Prompt)
	if err != nil {
		return err
	}

	newBoundaryPrompt, err := f(id+"_boundary_prompt", b.BoundaryPrompt)
	if err != nil {
		return err
	}

	b.Prompt, b.BoundaryPrompt = newPrompt, newBoundaryPrompt
	return nil
}

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

func (c *conversation) appendTurn(turn *models.Turn) {
	c.Lock()
	defer c.Unlock()

	c.history = append(c.history, turn)
}
