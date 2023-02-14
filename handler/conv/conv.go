package conv

import (
	"net/http"
	"strconv"

	"github.com/fox-one/pkg/httputil/param"
	"github.com/go-chi/chi"
	"github.com/pandodao/botastic/handler/render"
)

type (
	CreateConversationPayload struct {
		BotID uint64 `json:"bot_id"`
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

func CreateConversation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := &CreateConversationPayload{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		// @TODO create conversation

		render.JSON(w, nil)
	}
}

func GetConversation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, nil)
	}
}

func PostToConversation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conversationIDStr := chi.URLParam(r, "conversationID")
		conversationID, _ := strconv.ParseUint(conversationIDStr, 10, 64)
		if conversationID <= 0 {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		body := &PostToConversationPayload{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		// @TODO create conversation

		render.JSON(w, nil)
	}
}

func DeleteConversation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conversationIDStr := chi.URLParam(r, "conversationID")
		conversationID, _ := strconv.ParseUint(conversationIDStr, 10, 64)
		if conversationID <= 0 {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}
		// @TODO delete conversation
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
