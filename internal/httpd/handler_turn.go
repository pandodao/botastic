package httpd

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

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

	turn := &models.Turn{
		ConvID:  convID,
		Request: req.Content,
		Status:  api.TurnStatusInit,
	}

	if !h.createTurn(c, turn, false) {
		return
	}

	h.respData(c, api.CreateTurnResponse(turn.API()))
}

func (h *Handler) createTurn(c *gin.Context, turn *models.Turn, newConv bool) bool {
	if !newConv {
		// make sure no init turn exists in the conversation
		count, err := h.sh.GetTurnCount(c, turn.ConvID, api.TurnStatusInit)
		if err != nil {
			h.respErr(c, http.StatusInternalServerError, err)
			return false
		}
		if count != 0 {
			h.respErr(c, http.StatusBadRequest, errors.New("conversation already has an init turn"), api.ErrorCodeConversationHasInitTurn)
			return false
		}

		conv, err := h.sh.GetConv(c, turn.ConvID)
		if err != nil {
			h.respErr(c, http.StatusInternalServerError, err)
			return false
		}
		if conv == nil {
			h.respErr(c, http.StatusNotFound, errors.New("conversation not found"))
			return false
		}
		turn.BotID = conv.BotID
	}

	if err := h.sh.CreateTurn(c, turn); err != nil {
		h.respErr(c, http.StatusInternalServerError, err)
		return false
	}

	go func() {
		h.turnTransmitter.GetTurnsChan() <- turn
	}()

	return true
}

func (h *Handler) CreateTurnOneway(c *gin.Context) {
	var req api.CreateTurnOnewayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respErr(c, http.StatusBadRequest, err)
		return
	}

	var conv *models.Conv
	if req.ConversationID == uuid.Nil {
		conv = &models.Conv{
			BotID:        req.BotID,
			UserIdentity: req.UserIdentity,
		}
		if !h.createConv(c, conv) {
			return
		}
	} else {
		var err error
		conv, err = h.sh.GetConv(c, req.ConversationID)
		if err != nil {
			h.respErr(c, http.StatusInternalServerError, err)
			return
		}
		if conv == nil {
			h.respErr(c, http.StatusNotFound, errors.New("conversation not found"))
			return
		}
	}

	turn := &models.Turn{
		ConvID:  conv.ID,
		Request: req.Content,
		Status:  api.TurnStatusInit,
	}
	if conv != nil {
		turn.BotID = conv.BotID
	}

	if !h.createTurn(c, turn, conv != nil) {
		return
	}

	h.getTurn(c, turn, req.GetTurnRequest)
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

	h.getTurn(c, turn, req)
	return
}

func (h *Handler) getTurn(c *gin.Context, turn *models.Turn, req api.GetTurnRequest) {
	if turn.IsProcessed() || !req.BlockUntilProcessed {
		h.respData(c, api.GetTurnResponse(turn.API()))
		return
	}

	ctx := c.Request.Context()
	if req.TimeoutSeconds > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	if _, err := h.hub.AddAndWait(ctx, turn.ID); err != nil {
		h.respErr(c, http.StatusRequestTimeout, err)
		return
	}

	var err error
	turn, err = h.sh.GetTurn(c, turn.ID)
	if err != nil {
		h.respErr(c, http.StatusInternalServerError, err)
		return
	}

	h.respData(c, api.GetTurnResponse(turn.API()))
}
