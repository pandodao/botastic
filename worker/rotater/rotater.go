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
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/fox-one/pkg/logger"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/internal/chanhub"
	"github.com/pandodao/botastic/internal/gpt"
	"github.com/pandodao/botastic/internal/tiktoken"
	"github.com/pandodao/botastic/session"
	"github.com/pandodao/botastic/store"
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

		turnChan chan *core.ConvTurn
		hub      *chanhub.Hub

		processingMap   sync.Map
		tiktokenHandler *tiktoken.Handler
	}

	turnRequest struct {
		Conv       *core.Conversation
		Turn       *core.ConvTurn
		Bot        *core.Bot
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
	tiktokenHandler *tiktoken.Handler,
) *Worker {
	turnChan := make(chan *core.ConvTurn, MAX_REQUESTS_PER_MINUTE)
	return &Worker{
		cfg:             cfg,
		gptHandler:      gptHandler,
		convs:           convs,
		apps:            apps,
		models:          models,
		convz:           convz,
		botz:            botz,
		middlewarez:     middlewarez,
		userz:           userz,
		turnChan:        turnChan,
		hub:             hub,
		tiktokenHandler: tiktokenHandler,
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
		w.turnChan <- turn
	}
	return nil
}

func (w *Worker) UpdateConvTurnAsError(ctx context.Context, id uint64, mrs core.MiddlewareResults, err error) error {
	fmt.Printf("err: %v, %d\n", err, id)
	var tpe *core.TurnProcessError
	if !errors.As(err, &tpe) {
		tpe = core.NewTurnProcessError(core.TurnProcessErrorInternal, err)
	}
	if err := w.convs.UpdateConvTurn(ctx, id, "", 0, 0, 0, core.ConvTurnStatusError, mrs, tpe); err != nil {
		return err
	}
	return nil
}

func (w *Worker) ProcessConvTurn(ctx context.Context, turn *core.ConvTurn) error {
	log := logger.FromContext(ctx)

	fromMiddleware := turn.ID == 0
	if !fromMiddleware {
		if _, loaded := w.processingMap.LoadOrStore(turn.ID, struct{}{}); loaded {
			return nil
		}
		defer w.processingMap.Delete(turn.ID)
	}

	var (
		middlewareResults core.MiddlewareResults
		rr                *requestResult
		model             *core.Model
	)
	err := func() error {
		bot, err := w.botz.GetBot(ctx, turn.BotID)
		if err != nil {
			if store.IsNotFoundErr(err) {
				return core.NewTurnProcessError(core.TurnProcessBotNotFound, err)
			}
			return err
		}

		conv := &core.Conversation{}
		if turn.ConversationID != "" {
			conv, err = w.convz.GetConversation(ctx, turn.ConversationID)
			if err != nil {
				if store.IsNotFoundErr(err) {
					return core.NewTurnProcessError(core.TurnProcessConversationNotFound, err)
				}
				return err
			}
		}

		model, err = w.models.GetModel(ctx, bot.Model)
		if err != nil {
			if store.IsNotFoundErr(err) {
				return core.NewTurnProcessError(core.TurnProcessModelNotFound, err)
			}
			return err
		}

		middlewareCfg := bot.MiddlewareJson
		if turn.BotOverride.Middlewares != nil {
			middlewareCfg = *turn.BotOverride.Middlewares
		}

		additional := map[string]interface{}{}
		if len(middlewareCfg.Items) != 0 {
			app, err := w.apps.GetApp(ctx, turn.AppID)
			if err != nil {
				return core.NewTurnProcessError(core.TurnProcessMiddlewareError, err)
			}

			ctx = session.WithApp(ctx, app)
			middlewareResults = w.middlewarez.ProcessByConfig(ctx, middlewareCfg, turn)
			lastResult := middlewareResults[len(middlewareResults)-1]
			if lastResult.Err != nil && lastResult.Required {
				return core.NewTurnProcessError(core.TurnProcessMiddlewareError, lastResult.Err)
			}

			middlewareOutputs := make([]string, 0)
			for _, r := range middlewareResults {
				name := strings.ToUpper(strings.ReplaceAll(r.Name, "-", "_"))
				additional[fmt.Sprintf("MiddlewareOutput_%s", name)] = r.Result
				middlewareOutputs = append(middlewareOutputs, r.Result)
			}
			additional["MiddlewareOutput"] = strings.Join(middlewareOutputs, "\n\n")
		}

		turnReq := turnRequest{
			Conv:       conv,
			Turn:       turn,
			Bot:        bot,
			Model:      model,
			Additional: additional,
		}
		switch model.Provider {
		case core.ModelProviderOpenAI:
			rr, err = w.handleOpenAIProvider(ctx, turnReq)
		case core.ModelProviderCustom:
			rr, err = w.handleCustomProvider(ctx, turnReq)
		default:
			err = core.NewTurnProcessError(core.TurnProcessModelConfigError, fmt.Errorf("unsupported model provider: %s", model.Provider))
		}

		return err
	}()

	turn.Status = core.ConvTurnStatusCompleted
	turn.MiddlewareResults = middlewareResults
	turn.Response = rr.respText
	turn.PromptTokens = int(rr.promptTokenCount)
	turn.CompletionTokens = int(rr.completionTokenCount)
	turn.TotalTokens = int(rr.totalTokens)

	if !fromMiddleware {
		if err != nil {
			return w.UpdateConvTurnAsError(ctx, turn.ID, middlewareResults, err)
		}

		if err := w.convs.UpdateConvTurn(ctx, turn.ID, rr.respText, rr.promptTokenCount, rr.completionTokenCount, rr.totalTokens, core.ConvTurnStatusCompleted, middlewareResults, nil); err != nil {
			return err
		}
	}

	if err := w.userz.ConsumeCreditsByModel(ctx, turn.UserID, *model, rr.promptTokenCount, rr.completionTokenCount); err != nil {
		log.WithError(err).Warningf("userz.ConsumeCreditsByModel: model=%v, token=%v", model.Name(), rr.totalTokens)
	}

	if !fromMiddleware {
		// notify http handler
		w.hub.Broadcast(turn.ID, struct{}{})
		log.Printf("subwork processed a turn: %+v\n", turn.ID)
	}

	return nil
}

func (w *Worker) subworker(ctx context.Context, id int) {
	log := logger.FromContext(ctx).WithField("worker", "rotater.subworker").WithField("id", id)
	ctx = logger.WithContext(ctx, log)
	for {
		// Wait for a request to handle.
		turnReq := <-w.turnChan
		if err := w.ProcessConvTurn(ctx, turnReq); err != nil {
			log.WithError(err).Error("ProcessConvTurn")
		}
	}
}

type requestResult struct {
	respText             string
	totalTokens          int64
	promptTokenCount     int64
	completionTokenCount int64
}

func (w *Worker) handleCustomProvider(ctx context.Context, turnReq turnRequest) (*requestResult, error) {
	model := turnReq.Model
	conv := turnReq.Conv
	turn := turnReq.Turn
	bot := turnReq.Bot

	cc := model.CustomConfig
	if cc.Request.URL == "" {
		return nil, core.NewTurnProcessError(core.TurnProcessModelConfigError, fmt.Errorf("custom config request is empty"))
	}

	req, err := http.NewRequestWithContext(ctx, cc.Request.Method, cc.Request.URL, nil)
	if err != nil {
		return nil, core.NewTurnProcessError(core.TurnProcessModelConfigError, fmt.Errorf("invalid request url: %w", err))
	}

	prompt := bot.GetRequestContent(conv, turn.Request, turnReq.Additional)
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
			return nil, core.NewTurnProcessError(core.TurnProcessModelConfigError, fmt.Errorf("invalid request data: %w", err))
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
		return nil, core.NewTurnProcessError(core.TurnProcessModelCallError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, core.NewTurnProcessError(core.TurnProcessModelCallError, fmt.Errorf("Request failed with status code %d", resp.StatusCode))
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, core.NewTurnProcessError(core.TurnProcessModelCallError, fmt.Errorf("read response body error: %w", err))
	}

	respText := string(respData)
	if cc.Response.Path != "" {
		// get by gjson
		respText = gjson.Get(respText, cc.Response.Path).String()
	}

	promptTokenCount, _ := w.tiktokenHandler.CalToken(core.ModelProviderCustom, "", prompt)
	completionTokenCount, _ := w.tiktokenHandler.CalToken(core.ModelProviderCustom, "", respText)
	rr := &requestResult{
		respText:             respText,
		promptTokenCount:     int64(promptTokenCount),
		completionTokenCount: int64(completionTokenCount),
	}
	rr.totalTokens = rr.promptTokenCount + rr.completionTokenCount
	return rr, nil
}

func (w *Worker) handleOpenAIProvider(ctx context.Context, req turnRequest) (*requestResult, error) {
	model := req.Model
	conv := req.Conv
	turn := req.Turn
	bot := req.Bot

	var (
		chatRequest       *gogpt.ChatCompletionRequest
		completionRequest *gogpt.CompletionRequest
	)

	temperature := bot.Temperature
	if turn.BotOverride.Temperature != nil && *turn.BotOverride.Temperature >= 0 && *turn.BotOverride.Temperature <= 2 {
		temperature = *turn.BotOverride.Temperature
	}

	if model.IsOpenAIChatModel() {
		chatRequest = &gogpt.ChatCompletionRequest{
			Model:       model.ProviderModel,
			Messages:    req.Bot.GetChatMessages(req.Conv, req.Additional),
			Temperature: temperature,
		}
	} else if model.IsOpenAICompletionModel() {
		prompt := bot.GetRequestContent(conv, turn.Request, req.Additional)
		completionRequest = &gogpt.CompletionRequest{
			Model:       model.ProviderModel,
			Prompt:      prompt,
			Temperature: temperature,
			Stop:        []string{"Q:"},
			User:        conv.GetKey(),
		}
	} else {
		return nil, core.NewTurnProcessError(core.TurnProcessModelConfigError, fmt.Errorf("model %s is not supported", model.Name()))
	}

	var usage gogpt.Usage
	var respText string
	if completionRequest != nil {
		gptResp, err := w.gptHandler.CreateCompletion(ctx, *completionRequest)
		if err != nil {
			return nil, core.NewTurnProcessError(core.TurnProcessModelCallError, err)
		}
		prefix := "A:"
		respText = strings.TrimPrefix(gptResp.Choices[0].Text, prefix)
		respText = strings.TrimSpace(respText)
		usage = gptResp.Usage
	} else {
		gptResp, err := w.gptHandler.CreateChatCompletion(ctx, *chatRequest)
		if err != nil {
			return nil, core.NewTurnProcessError(core.TurnProcessModelCallError, err)
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
