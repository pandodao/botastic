package conv

import (
	"net/http"
	"strconv"

	"github.com/fox-one/pkg/httputil/param"
	"github.com/go-chi/chi"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/render"
	"github.com/pandodao/botastic/session"
)

type (
	CreateConversationPayload struct {
		BotID        uint64 `json:"bot_id"`
		UserIdentity string `json:"user_identity"`
		UpdateConversationPayload
	}

	UpdateConversationPayload struct {
		Lang string `json:"lang"`
	}

	PostToConversationPayload struct {
		Content  string `json:"content"`
		Category string `json:"category"`
	}
)

func CreateConversation(botz core.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		app := session.AppFrom(ctx)
		if app == nil {
			render.Error(w, http.StatusUnauthorized, nil)
			return
		}

		body := &CreateConversationPayload{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		if body.BotID <= 0 || body.UserIdentity == "" || body.Lang == "" {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		conv, err := botz.CreateConversation(ctx, body.BotID, app.ID, body.UserIdentity, body.Lang)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, conv)
	}
}

func GetConversation(botz core.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		app := session.AppFrom(ctx)
		if app == nil {
			render.Error(w, http.StatusUnauthorized, nil)
			return
		}

		conversationID := chi.URLParam(r, "conversationID")
		if conversationID == "" {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		conv, err := botz.GetConversation(ctx, conversationID)
		if err != nil {
			render.Error(w, http.StatusNotFound, err)
			return
		}

		if conv.App.ID != app.ID {
			render.Error(w, http.StatusNotFound, nil)
			return
		}

		render.JSON(w, conv)
	}
}

func PostToConversation(botz core.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		app := session.AppFrom(ctx)
		if app == nil {
			render.Error(w, http.StatusUnauthorized, nil)
			return
		}

		conversationID := chi.URLParam(r, "conversationID")

		body := &PostToConversationPayload{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		conv, err := botz.GetConversation(ctx, conversationID)
		if err != nil {
			render.Error(w, http.StatusNotFound, nil)
		}

		if conv.App.ID != app.ID {
			render.Error(w, http.StatusNotFound, nil)
			return
		}

		// @TODO post to conversation

		render.JSON(w, nil)
	}
}

func DeleteConversation(botz core.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		app := session.AppFrom(ctx)
		if app == nil {
			render.Error(w, http.StatusUnauthorized, nil)
			return
		}

		conversationID := chi.URLParam(r, "conversationID")
		conv, err := botz.GetConversation(ctx, conversationID)
		if err != nil {
			render.Error(w, http.StatusNotFound, nil)
		}

		if conv.App.ID != app.ID {
			render.Error(w, http.StatusNotFound, nil)
			return
		}

		botz.DeleteConversation(ctx, conversationID)

		render.JSON(w, nil)
	}
}

func UpdateConversation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conversationIDStr := chi.URLParam(r, "conversationID")
		conversationID, _ := strconv.ParseUint(conversationIDStr, 10, 64)
		if conversationID <= 0 {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		body := &UpdateConversationPayload{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		// @TODO update conversation
		render.JSON(w, nil)
	}
}
