package bot

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/fox-one/pkg/httputil/param"
	"github.com/go-chi/chi"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/render"
	"github.com/pandodao/botastic/session"
)

type CreateOrUpdateBotPayload struct {
	Name             string                `json:"name"`
	Model            string                `json:"model"`
	Prompt           string                `json:"prompt"`
	Temperature      float32               `json:"temperature"`
	MaxTurnCount     int                   `json:"max_turn_count"`
	ContextTurnCount int                   `json:"context_turn_count"`
	Middlewares      core.MiddlewareConfig `json:"middlewares"`
}

func (body *CreateOrUpdateBotPayload) Formalize(defaultValue *core.Bot) error {
	body.Name = strings.TrimSpace(body.Name)
	body.Model = strings.TrimSpace(body.Model)
	if len(body.Name) > 128 || len(body.Name) == 0 {
		return core.ErrBotIncorrectField
	}

	if len(body.Model) > 128 || len(body.Model) == 0 {
		return core.ErrBotIncorrectField
	}

	if defaultValue != nil {
		if body.Temperature <= 0 {
			body.Temperature = defaultValue.Temperature
		}
		if body.MaxTurnCount <= 0 {
			body.MaxTurnCount = defaultValue.MaxTurnCount
		}
		if body.ContextTurnCount <= 0 {
			body.ContextTurnCount = defaultValue.ContextTurnCount
		}
		if body.Middlewares.Items == nil {
			body.Middlewares.Items = defaultValue.Middlewares.Items
		}
	} else {
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
	}

	return nil
}

func GetBot(botz core.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, found := session.UserFrom(ctx)
		if !found {
			render.Error(w, http.StatusUnauthorized, core.ErrUnauthorized)
			return
		}

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

		if bot.UserID != user.ID {
			render.Error(w, http.StatusNotFound, core.ErrBotNotFound)
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

func GetMyBots(botz core.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, found := session.UserFrom(ctx)
		if !found {
			render.Error(w, http.StatusUnauthorized, core.ErrUnauthorized)
			return
		}

		bots, err := botz.GetBotsByUserID(ctx, user.ID)
		if err != nil {
			render.JSON(w, []interface{}{})
			return
		}

		render.JSON(w, bots)
	}
}

func UpdateBot(botz core.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, found := session.UserFrom(ctx)
		if !found {
			render.Error(w, http.StatusUnauthorized, core.ErrUnauthorized)
			return
		}

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

		if bot.UserID != user.ID {
			render.Error(w, http.StatusNotFound, core.ErrBotNotFound)
			return
		}

		body := &CreateOrUpdateBotPayload{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		if err := body.Formalize(bot); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		err = botz.UpdateBot(ctx, botID, body.Name, body.Model, body.Prompt, body.Temperature, body.MaxTurnCount, body.ContextTurnCount, body.Middlewares, false)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, bot)
	}
}

func CreateBot(botz core.BotService, models core.ModelStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, found := session.UserFrom(ctx)
		if !found {
			render.Error(w, http.StatusUnauthorized, core.ErrUnauthorized)
			return
		}

		botArr, _ := botz.GetBotsByUserID(ctx, user.ID)
		if len(botArr) >= 3 {
			render.Error(w, http.StatusBadRequest, core.ErrAppLimitReached)
			return
		}

		body := &CreateOrUpdateBotPayload{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		if err := body.Formalize(nil); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		bot, err := botz.CreateBot(ctx, user.ID, body.Name, body.Model, body.Prompt, body.Temperature, body.MaxTurnCount, body.ContextTurnCount, body.Middlewares, false)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if errors.Is(err, core.ErrInvalidModel) {
				statusCode = http.StatusBadRequest
			}
			render.Error(w, statusCode, err)
			return
		}

		render.JSON(w, bot)
	}
}

func DeleteBot(botz core.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, found := session.UserFrom(ctx)
		if !found {
			render.Error(w, http.StatusUnauthorized, core.ErrUnauthorized)
		}

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

		if bot.UserID != user.ID {
			render.Error(w, http.StatusNotFound, core.ErrBotNotFound)
			return
		}

		if err := botz.DeleteBot(ctx, bot.ID); err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, bot)
	}
}
