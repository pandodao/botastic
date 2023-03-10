package bot

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/fox-one/pkg/httputil/param"
	"github.com/go-chi/chi"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/render"
	"github.com/pandodao/botastic/session"
)

type CreateBotPayload struct {
	Name             string                `json:"name"`
	Model            string                `json:"model"`
	Prompt           string                `json:"prompt"`
	Temperature      float32               `json:"temperature"`
	MaxTurnCount     int                   `json:"max_turn_count"`
	ContextTurnCount int                   `json:"context_turn_count"`
	Middlewares      core.MiddlewareConfig `json:"middlewares"`
	Public           bool                  `json:"public"`
}

func GetBot(botz core.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		botIDStr := chi.URLParam(r, "botID")
		botID, _ := strconv.ParseUint(botIDStr, 10, 64)

		if botID <= 0 {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		bot, err := botz.GetBot(ctx, botID)
		if err != nil {
			render.Error(w, http.StatusNotFound, err)
			return
		}

		render.JSON(w, bot)
	}
}

func GetPublicBots(botz core.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bots, err := botz.GetPublicBots(ctx)
		if err != nil {
			render.JSON(w, []interface{}{})
			return
		}
		render.JSON(w, bots)
	}
}

func CreateBot(botz core.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, found := session.UserFrom(ctx)
		if !found {
			render.Error(w, http.StatusUnauthorized, core.ErrUnauthorized)
			return
		}

		body := &CreateBotPayload{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		body.Name = strings.TrimSpace(body.Name)

		botArr, _ := botz.GetBotsByUserID(ctx, user.ID)
		if len(botArr) >= 10 {
			render.Error(w, http.StatusBadRequest, core.ErrAppLimitReached)
			return
		}

		if len(body.Name) > 128 || len(body.Name) == 0 {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		// @TODO verify model name

		if body.Temperature <= 0 {
			body.Temperature = 1
		}

		if body.MaxTurnCount <= 0 {
			body.MaxTurnCount = 4
		}

		if body.ContextTurnCount <= 0 {
			body.ContextTurnCount = 4
		}

		if body.Middlewares.Items == nil {
			body.Middlewares.Items = []*core.Middleware{}
		}

		bot, err := botz.CreateBot(ctx, user.ID, body.Name, body.Model, body.Prompt, body.Temperature, body.MaxTurnCount, body.ContextTurnCount, body.Middlewares, body.Public)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, bot)
	}
}
