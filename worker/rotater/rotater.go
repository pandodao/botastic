package rotater

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/fox-one/pkg/logger"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/internal/chanhub"
	"github.com/pandodao/botastic/internal/gpt"
	"github.com/pandodao/botastic/session"
	"github.com/pandodao/tokenizer-go"
	gogpt "github.com/sashabaranov/go-openai"
	"github.com/tidwall/gjson"
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
		Conv       *core.Conversation
		Turn       *core.ConvTurn
		Model      *core.Model
		Additional map[string]any
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
	processingIds := make([]uint64, 0)
	w.processingMap.Range(func(key, value interface{}) bool {
		processingIds = append(processingIds, key.(uint64))
		return true
	})

	turns, err := w.convs.GetConvTurnsByStatus(ctx, processingIds, []int{core.ConvTurnStatusInit, core.ConvTurnStatusPending})
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

		model, err := w.models.GetModel(ctx, conv.Bot.Model)
		if err != nil {
			w.UpdateConvTurnAsError(ctx, turn.ID, fmt.Errorf("unsupported model: %s", conv.Bot.Model).Error())
			continue
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

		if turn.Status == core.ConvTurnStatusInit {
			if err := w.convs.UpdateConvTurn(ctx, turn.ID, "", 0, 0, 0, core.ConvTurnStatusPending); err != nil {
				continue
			}
		}

		w.turnReqChan <- TurnRequest{
			Conv:       conv,
			Turn:       turn,
			Model:      model,
			Additional: additional,
		}
	}
	return nil
}

func (w *Worker) UpdateConvTurnAsError(ctx context.Context, id uint64, errMsg string) error {
	fmt.Printf("errMsg: %v, %d\n", errMsg, id)
	if err := w.convs.UpdateConvTurn(ctx, id, "Something wrong happened", 0, 0, 0, core.ConvTurnStatusError); err != nil {
		return err
	}
	return nil
}

func (w *Worker) subworker(ctx context.Context, id int) {
	log := logger.FromContext(ctx).WithField("worker", "rotater.subworker")
	for {
		// Wait for a request to handle.
		turnReq := <-w.turnReqChan
		turn := turnReq.Turn

		func() {
			if _, loaded := w.processingMap.LoadOrStore(turn.ID, struct{}{}); loaded {
				return
			}
			defer w.processingMap.Delete(turn.ID)

			var (
				rr  *requestResult
				err error
			)
			switch turnReq.Model.Provider {
			case core.ModelProviderOpenAI:
				rr, err = w.handleOpenAIProvider(ctx, turnReq)
			case core.ModelProviderCustom:
				rr, err = w.handleCustomProvider(ctx, turnReq)
			}

			if err != nil {
				w.UpdateConvTurnAsError(ctx, turn.ID, err.Error())
				return
			}

			if err := w.convs.UpdateConvTurn(ctx, turn.ID, rr.respText, rr.promptTokenCount, rr.completionTokenCount, rr.totalTokens, core.ConvTurnStatusCompleted); err != nil {
				return
			}

			if err := w.userz.ConsumeCreditsByModel(ctx, turn.UserID, *turnReq.Model, rr.promptTokenCount, rr.completionTokenCount); err != nil {
				log.WithError(err).Warningf("userz.ConsumeCreditsByModel: model=%v, token=%v", turnReq.Model.Name(), rr.totalTokens)
			}

			// notify http handler
			w.hub.Broadcast(strconv.FormatUint(turn.ID, 10), struct{}{})

			log.Printf("subwork-%03d processed a turn: %+v\n", id, turn.ID)
		}()
	}
}

type requestResult struct {
	respText             string
	totalTokens          int64
	promptTokenCount     int64
	completionTokenCount int64
}

func (w *Worker) handleCustomProvider(ctx context.Context, turnReq TurnRequest) (*requestResult, error) {
	model := turnReq.Model
	conv := turnReq.Conv
	turn := turnReq.Turn
	bot := conv.Bot

	cc, err := model.UnmarshalCustomConfig()
	if err != nil {
		return nil, fmt.Errorf("unmarshal custom config error: %v", err)
	}

	if cc.Request.URL == "" {
		return nil, fmt.Errorf("custom config request is empty")
	}

	req, err := http.NewRequestWithContext(ctx, cc.Request.Method, cc.Request.URL, nil)
	if err != nil {
		return nil, err
	}

	prompt := bot.GetPrompt(conv, turn.Request, turnReq.Additional)
	for key, value := range cc.Request.Data {
		vs, ok := value.(string)
		if !ok {
			continue
		}
		t, err := template.New(fmt.Sprintf("%d-data-%s", bot.ID, key)).Parse(vs)
		if err != nil {
			continue
		}

		buf := new(bytes.Buffer)
		if err := t.Execute(buf, map[string]any{
			"PROMPT": prompt,
		}); err != nil {
			continue
		}

		cc.Request.Data[key] = buf.String()
	}

	// If the request method is POST, set the request body to the JSON-encoded data
	if cc.Request.Method == "POST" {
		jsonData, err := json.Marshal(cc.Request.Data)
		if err != nil {
			return nil, err
		}
		req.Body = ioutil.NopCloser(bytes.NewReader(jsonData))
		req.ContentLength = int64(len(jsonData))
		req.Header.Set("Content-Type", "application/json")
	} else if cc.Request.Method == "GET" {
		// If the request method is GET, encode the data as query string parameters
		queryParams := url.Values{}
		for key, value := range cc.Request.Data {
			queryParams.Set(key, fmt.Sprintf("%v", value))
		}
		req.URL.RawQuery = queryParams.Encode()
	}

	for k, v := range cc.Request.Headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Request failed with status code %d", resp.StatusCode))
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	respText := string(respData)
	if cc.Response.Path != "" {
		// get by gjson
		respText = gjson.Get(respText, cc.Response.Path).String()
	}

	rr := &requestResult{
		respText:             respText,
		promptTokenCount:     tokenizer.MustCalToken(prompt),
		completionTokenCount: tokenizer.MustCalToken(respText),
	}
	rr.totalTokens = rr.promptTokenCount + rr.completionTokenCount
	return rr, nil
}

func (w *Worker) handleOpenAIProvider(ctx context.Context, req TurnRequest) (*requestResult, error) {
	model := req.Model
	conv := req.Conv
	turn := req.Turn
	bot := conv.Bot

	var (
		chatRequest       *gogpt.ChatCompletionRequest
		completionRequest *gogpt.CompletionRequest
	)

	if model.IsOpenAIChatModel() {
		chatRequest = &gogpt.ChatCompletionRequest{
			Model:       model.ProviderModel,
			Messages:    req.Conv.Bot.GetChatMessages(req.Conv, req.Additional),
			Temperature: bot.Temperature,
		}
	} else if model.IsOpenAICompletionModel() {
		prompt := bot.GetPrompt(conv, turn.Request, req.Additional)
		completionRequest = &gogpt.CompletionRequest{
			Model:       model.ProviderModel,
			Prompt:      prompt,
			Temperature: bot.Temperature,
			Stop:        []string{"Q:"},
			User:        conv.GetKey(),
		}
	} else {
		return nil, core.ErrInvalidModel
	}

	var usage gogpt.Usage
	var respText string
	if completionRequest != nil {
		gptResp, err := w.gptHandler.CreateCompletion(ctx, *completionRequest)
		if err != nil {
			return nil, err
		}
		prefix := "A:"
		respText = strings.TrimPrefix(gptResp.Choices[0].Text, prefix)
		respText = strings.TrimSpace(respText)
		usage = gptResp.Usage
	} else {
		gptResp, err := w.gptHandler.CreateChatCompletion(ctx, *chatRequest)
		if err != nil {
			return nil, err
		}
		respText = strings.TrimSpace(gptResp.Choices[0].Message.Content)
		usage = gptResp.Usage
	}

	return &requestResult{
		respText:             respText,
		totalTokens:          int64(usage.TotalTokens),
		promptTokenCount:     int64(usage.PromptTokens),
		completionTokenCount: int64(usage.CompletionTokens),
	}, nil
}
