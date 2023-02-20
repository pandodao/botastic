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

func CreateConversation(botz core.BotService, convz core.ConversationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		app := session.AppFrom(ctx)

		body := &CreateConversationPayload{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		if body.BotID <= 0 || body.UserIdentity == "" || body.Lang == "" {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		conv, err := convz.CreateConversation(ctx, body.BotID, app.ID, body.UserIdentity, body.Lang)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, conv)
	}
}

func GetConversation(botz core.BotService, convz core.ConversationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		app := session.AppFrom(ctx)

		conversationID := chi.URLParam(r, "conversationID")
		if conversationID == "" {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		conv, err := convz.GetConversation(ctx, conversationID)
		if err != nil || conv == nil {
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

func PostToConversation(botz core.BotService, convz core.ConversationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		app := session.AppFrom(ctx)

		conversationID := chi.URLParam(r, "conversationID")

		body := &PostToConversationPayload{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		conv, err := convz.GetConversation(ctx, conversationID)
		if err != nil || conv == nil {
			render.Error(w, http.StatusNotFound, nil)
			return
		}

		if conv.App.ID != app.ID {
			render.Error(w, http.StatusNotFound, nil)
			return
		}

		// @TODO post to conversation
		turn, err := convz.PostToConversation(ctx, conv, body.Content)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, turn)
	}
}

func DeleteConversation(botz core.BotService, convz core.ConversationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		app := session.AppFrom(ctx)

		conversationID := chi.URLParam(r, "conversationID")
		conv, err := convz.GetConversation(ctx, conversationID)
		if err != nil || conv == nil {
			render.Error(w, http.StatusNotFound, nil)
			return
		}

		if conv.App.ID != app.ID {
			render.Error(w, http.StatusNotFound, nil)
			return
		}

		convz.DeleteConversation(ctx, conversationID)

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
