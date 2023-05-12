package httpd

import (
	"context"
	"errors"
	"net/http"
	"strconv"

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

func (h *Handler) CreateTurnOneway(c *gin.Context) {
	var req api.CreateTurnOnewayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respErr(c, http.StatusBadRequest, err)
		return
	}

	var (
		conv       *models.Conv
		err        error
		newCreated bool
	)
	if req.ConversationID == uuid.Nil {
		newCreated = true
		conv = &models.Conv{
			BotID:        req.CreateConvRequest.BotID,
			UserIdentity: req.CreateConvRequest.UserIdentity,
		}
		err = h.sh.CreateConv(c, conv)
	} else {
		conv, err = h.sh.GetConv(c, req.ConversationID)
	}
	if err != nil {
		h.respErr(c, http.StatusInternalServerError, err)
		return
	}

	if !newCreated {
		// make sure no init turn exists in the conversation
		count, err := h.sh.GetTurnCount(c, conv.ID, api.TurnStatusInit)
		if err != nil {
			h.respErr(c, http.StatusInternalServerError, err)
			return
		}
		if count != 0 {
			h.respErr(c, http.StatusBadRequest, errors.New("conversation already has an init turn"), api.ErrorCodeConversationHasInitTurn)
			return
		}
	}

	turn := &models.Turn{
		ConvID:  conv.ID,
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

func (h *Handler) GetTurn(c *gin.Context) {
	turnIDStr := c.Param("turn_id")
	turnID, err := strconv.ParseUint(turnIDStr, 10, 64)
	if err != nil {
		h.respErr(c, http.StatusBadRequest, err)
		return
	}

	var req api.GetTurnRequest
	if err := c.Bind(&req); err != nil {
		h.respErr(c, http.StatusBadRequest, err)
		return
	}

	turn, err := h.sh.GetTurn(c, uint(turnID))
	if err != nil {
		h.respErr(c, http.StatusInternalServerError, err)
		return
	}
	if turn == nil {
		h.respErr(c, http.StatusNotFound, errors.New("turn not found"))
		return
	}
	if turn.IsProcessed() || !req.BlockUntilProcessed {
		h.respData(c, api.GetTurnResponse(turn.API()))
		return
	}

	ctx := c.Request.Context()
	if req.Timeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	if _, err := h.hub.AddAndWait(ctx, turnID); err != nil {
		h.respErr(c, http.StatusRequestTimeout, err)
		return
	}

	turn, err = h.sh.GetTurn(c, uint(turnID))
	if err != nil {
		h.respErr(c, http.StatusInternalServerError, err)
		return
	}

	h.respData(c, api.GetTurnResponse(turn.API()))
}
