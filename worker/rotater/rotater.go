package rotater

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fox-one/pkg/logger"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/internal/chanhub"
	"github.com/pandodao/botastic/internal/gpt"
	"github.com/pandodao/botastic/session"
	gogpt "github.com/sashabaranov/go-openai"
)

const (
	MAX_SUBWORKERS          = 32
	MAX_REQUESTS_PER_MINUTE = 3000
)

type (
	Config struct {
	}

	Worker struct {
		cfg        Config
		gptHandler *gpt.Handler
		convs      core.ConversationStore
		apps       core.AppStore
		models     core.ModelStore

		convz       core.ConversationService
		botz        core.BotService
		middlewarez core.MiddlewareService
		userz       core.UserService

		turnReqChan chan TurnRequest
		hub         *chanhub.Hub

		processingMap sync.Map
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
	apps core.AppStore,
	models core.ModelStore,

	convz core.ConversationService,
	botz core.BotService,
	middlewarez core.MiddlewareService,
	userz core.UserService,

	hub *chanhub.Hub,
) *Worker {
	turnReqChan := make(chan TurnRequest, MAX_REQUESTS_PER_MINUTE)
	return &Worker{
		cfg:         cfg,
		gptHandler:  gptHandler,
		convs:       convs,
		apps:        apps,
		models:      models,
		convz:       convz,
		botz:        botz,
		middlewarez: middlewarez,
		userz:       userz,

		turnReqChan: turnReqChan,
		hub:         hub,
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
	processinngIds := make([]uint64, 0)
	w.processingMap.Range(func(key, value interface{}) bool {
		processinngIds = append(processinngIds, key.(uint64))
		return true
	})

	turns, err := w.convs.GetConvTurnsByStatus(ctx, processinngIds, []int{core.ConvTurnStatusInit, core.ConvTurnStatusPending})
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

		additional := map[string]interface{}{}
		if bot.Middlewares.Items != nil && len(bot.Middlewares.Items) != 0 {
			middlewareOutputs := make([]string, 0)
			app, err := w.apps.GetApp(ctx, turn.AppID)
			if err == nil {
				for _, middleware := range bot.Middlewares.Items {
					ctx = session.WithApp(ctx, app)
					result, err := w.middlewarez.Process(ctx, middleware, turn.Request)
					if err == nil && result != nil {
						middlewareOutputs = append(middlewareOutputs, result.Result)
					}
				}
			}
			additional["MiddlewareOutput"] = strings.Join(middlewareOutputs, "\n\n")
		}

		model, err := w.models.GetModel(ctx, bot.Model)
		if err != nil {
			w.UpdateConvTurnAsError(ctx, turn.ID, fmt.Errorf("unsupported model: %s", bot.Model).Error())
			continue
		}

		if model.Props.IsOpenAIChatModel {
			request := gogpt.ChatCompletionRequest{
				Model:       model.ProviderModel,
				Messages:    bot.GetChatMessages(conv, additional),
				Temperature: bot.Temperature,
			}

			turnReq.ChatRequest = &request

		} else if model.Props.IsOpenAICompletionModel {
			prompt := bot.GetPrompt(conv, turn.Request)
			request := gogpt.CompletionRequest{
				Model:       model.ProviderModel,
				Prompt:      prompt,
				Temperature: bot.Temperature,
				Stop:        []string{"Q:"},
				User:        conv.GetKey(),
			}
			turnReq.Request = &request

		} else {
			w.UpdateConvTurnAsError(ctx, turn.ID, core.ErrInvalidModel.Error())
			continue
		}

		if turn.Status == core.ConvTurnStatusInit {
			if err := w.convs.UpdateConvTurn(ctx, turn.ID, "", 0, core.ConvTurnStatusPending); err != nil {
				continue
			}
		}

		w.turnReqChan <- turnReq
	}
	return nil
}

func (w *Worker) UpdateConvTurnAsError(ctx context.Context, id uint64, errMsg string) error {
	fmt.Printf("errMsg: %v, %d\n", errMsg, id)
	if err := w.convs.UpdateConvTurn(ctx, id, "Something wrong happened", 0, core.ConvTurnStatusError); err != nil {
		return err
	}
	return nil
}

func (w *Worker) subworker(ctx context.Context, id int) {
	log := logger.FromContext(ctx).WithField("worker", "rotater.subworker")
	for {
		// Wait for a request to handle.
		turnReq := <-w.turnReqChan

		func() {
			if _, loaded := w.processingMap.LoadOrStore(turnReq.TurnID, struct{}{}); loaded {
				return
			}
			defer w.processingMap.Delete(turnReq.TurnID)

			respText := ""
			var totalTokens int
			var promptTokenCount int
			var completionTokenCount int
			var err error
			switch {
			case turnReq.Request != nil:
				var gptResp gogpt.CompletionResponse
				gptResp, err = w.gptHandler.CreateCompletion(ctx, *turnReq.Request)
				if err == nil {
					prefix := "A:"
					respText = strings.TrimPrefix(gptResp.Choices[0].Text, prefix)
					respText = strings.TrimSpace(respText)
					totalTokens = gptResp.Usage.TotalTokens
					promptTokenCount = gptResp.Usage.PromptTokens
					completionTokenCount = gptResp.Usage.CompletionTokens
				}

			case turnReq.ChatRequest != nil:
				var gptResp gogpt.ChatCompletionResponse
				gptResp, err = w.gptHandler.CreateChatCompletion(ctx, *turnReq.ChatRequest)
				if err == nil {
					respText = strings.TrimSpace(gptResp.Choices[0].Message.Content)
					totalTokens = gptResp.Usage.TotalTokens
					promptTokenCount = gptResp.Usage.PromptTokens
					completionTokenCount = gptResp.Usage.CompletionTokens
				}
			}

			if err != nil {
				w.UpdateConvTurnAsError(ctx, turnReq.TurnID, err.Error())
				return
			}

			if err := w.convs.UpdateConvTurn(ctx, turnReq.TurnID, respText, totalTokens, core.ConvTurnStatusCompleted); err != nil {
				return
			}

			turn, err := w.convs.GetConvTurn(ctx, turnReq.TurnID)
			if err == nil {
				conv, err := w.convz.GetConversation(ctx, turn.ConversationID)
				if err != nil {
					log.WithError(err).Warningf("convz.GetConversation error")
				} else {
					if err := w.userz.ConsumeCreditsByModel(ctx, turn.UserID, conv.Bot.Model, int64(promptTokenCount), int64(completionTokenCount)); err != nil {
						log.WithError(err).Warningf("userz.ConsumeCreditsByModel: model=%v, token=%v", conv.Bot.Model, totalTokens)
					}
				}
			}

			// notify http handler
			w.hub.Broadcast(strconv.FormatUint(turnReq.TurnID, 10), struct{}{})

			log.Printf("subwork-%03d processed a turn: %+v\n", id, turnReq.TurnID)

		}()
	}
}
