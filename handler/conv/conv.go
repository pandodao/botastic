package conv

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/fox-one/pkg/httputil/param"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/render"
	"github.com/pandodao/botastic/internal/chanhub"
	"github.com/pandodao/botastic/session"
	"gorm.io/gorm"
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
		BotOverride core.BotOverride `json:"bot_override"`
		Content     string           `json:"content"`
		Category    string           `json:"category"`
	}

	CreateOnewayConversationPayload struct {
		BotID       uint64           `json:"bot_id"`
		BotOverride core.BotOverride `json:"bot_override"`
		Content     string           `json:"content"`
		Lang        string           `json:"lang"`
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

func GetConversationTurn(botz core.BotService, convs core.ConversationStore, hub *chanhub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		app := session.AppFrom(ctx)

		bur := r.URL.Query().Get("block_until_response")
		blockUntilResponse, _ := strconv.ParseBool(bur)

		conversationID := chi.URLParam(r, "conversationID")
		if conversationID == "" {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		turnIDStr := chi.URLParam(r, "turnID")
		if turnIDStr == "" {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		turnId, err := strconv.ParseUint(turnIDStr, 10, 64)
		if err != nil {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		convTurn, err := convs.GetConvTurn(ctx, turnId)
		if err != nil && err != gorm.ErrRecordNotFound {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		if convTurn == nil || convTurn.ID == 0 || convTurn.ConversationID != conversationID {
			render.Error(w, http.StatusBadRequest, fmt.Errorf("no conversation turn"))
			return
		}

		if convTurn.AppID != app.ID {
			render.Error(w, http.StatusBadRequest, fmt.Errorf("no conversation turn"))
			return
		}

		if !blockUntilResponse || convTurn.IsProcessed() {
			render.JSON(w, convTurn)
			return
		}

		_, err = hub.AddAndWait(ctx, turnIDStr)
		if err != nil {
			if err == context.Canceled {
				render.Error(w, http.StatusBadRequest, err)
				return
			}
		}
		convTurn, err = convs.GetConvTurn(ctx, turnId)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, convTurn)
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

		if body.BotOverride.Temperature != nil && (*body.BotOverride.Temperature < 0 || *body.BotOverride.Temperature > 2) {
			render.Error(w, http.StatusBadRequest, nil)
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

		// check if the conversation has pending turn
		if len(conv.History) > 0 && !conv.History[len(conv.History)-1].IsProcessed() {
			render.Error(w, http.StatusTooManyRequests, core.ErrConvTurnNotProcessed)
			return
		}

		turn, err := convz.PostToConversation(ctx, conv, body.Content, body.BotOverride)
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

func CreateOnewayConversation(convz core.ConversationService, convs core.ConversationStore, hub *chanhub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		app := session.AppFrom(ctx)

		body := &CreateOnewayConversationPayload{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		if body.BotID <= 0 {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		if body.BotOverride.Temperature != nil && (*body.BotOverride.Temperature < 0 || *body.BotOverride.Temperature > 2) {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		uid := uuid.New().String()

		conv, err := convz.CreateConversation(ctx, body.BotID, app.ID, uid, body.Lang)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		turn, err := convz.PostToConversation(ctx, conv, body.Content, body.BotOverride)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		turnIDStr := strconv.FormatUint(turn.ID, 10)

		_, err = hub.AddAndWait(ctx, turnIDStr)
		if err != nil {
			if err == context.Canceled {
				render.Error(w, http.StatusBadRequest, err)
				return
			}
		}

		turn, err = convs.GetConvTurn(ctx, turn.ID)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, turn)

	}
}
