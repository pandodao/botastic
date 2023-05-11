package httpd

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pandodao/botastic/api"
	"github.com/pandodao/botastic/models"
)

func (h *Handler) CreateTurn(c *gin.Context) {
	convIDStr := c.Param("conv_id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		h.respErr(c, http.StatusBadRequest, err)
		return
	}

	var req api.CreateTurnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respErr(c, http.StatusBadRequest, err)
		return
	}

	// make sure no init turn exists in the conversation
	count, err := h.sh.GetTurnCount(c, convID, api.TurnStatusInit)
	if err != nil {
		h.respErr(c, http.StatusInternalServerError, err)
		return
	}
	if count != 0 {
		h.respErr(c, http.StatusBadRequest, errors.New("conversation already has an init turn"), api.ErrorCodeConversationHasInitTurn)
		return
	}

	conv, err := h.sh.GetConv(c, convID)
	if err != nil {
		h.respErr(c, http.StatusInternalServerError, err)
		return
	}
	if conv == nil {
		h.respErr(c, http.StatusNotFound, errors.New("conv not found"))
		return
	}

	turn := &models.Turn{
		ConvID:  convID,
		BotID:   conv.BotID,
		Request: req.Content,
		Status:  api.TurnStatusInit,
	}

	if err := h.sh.CreateTurn(c, turn); err != nil {
		h.respErr(c, http.StatusInternalServerError, err)
		return
	}

	h.respData(c, api.CreateTurnResponse(turn.API()))
}
