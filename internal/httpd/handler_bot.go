package httpd

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pandodao/botastic/api"
	"github.com/pandodao/botastic/models"
)

func (h *Handler) CreateBot(c *gin.Context) {
	var req api.CreateBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respErr(c, http.StatusBadRequest, err)
		return
	}

	if _, err := h.llms.GetChatModel(req.ChatModel); err != nil {
		h.respErr(c, http.StatusBadRequest, errors.New("chat model does not exist"))
		return
	}
	if req.Middlewares != nil {
		if err := h.middlewareHandler.ValidateConfig(req.Middlewares); err != nil {
			h.respErr(c, http.StatusBadRequest, err)
			return
		}
	}

	bot := &models.Bot{
		Name:             req.Name,
		Prompt:           req.Prompt,
		ChatModel:        req.ChatModel,
		BoundaryPrompt:   req.BoundaryPrompt,
		ContextTurnCount: req.ContextTurnCount,
		Temperature:      req.Temperature,
	}
	if req.Middlewares != nil {
		v := models.MiddlewareConfig(*req.Middlewares)
		bot.Middlewares = &v
	}

	if err := h.sh.CreateBot(c, bot); err != nil {
		h.respErr(c, http.StatusInternalServerError, err)
		return
	}

	h.respData(c, api.CreateBotResponse(bot.API()))
}

func (h *Handler) GetBot(c *gin.Context) {
	botIDStr := c.Param("bot_id")
	botId, err := strconv.ParseUint(botIDStr, 10, 64)
	if err != nil {
		h.respErr(c, http.StatusBadRequest, err)
		return
	}

	bot, err := h.sh.GetBot(c, uint(botId))
	if err != nil {
		h.respErr(c, http.StatusInternalServerError, err)
		return
	}
	if bot == nil {
		h.respErr(c, http.StatusNotFound, err)
		return
	}

	h.respData(c, api.GetBotResponse(bot.API()))
}

func (h *Handler) GetBots(c *gin.Context) {
	bots, err := h.sh.GetBots(c)
	if err != nil {
		h.respErr(c, http.StatusInternalServerError, err)
		return
	}

	data := make(api.GetBotsResponse, 0, len(bots))
	for _, b := range bots {
		data = append(data, b.API())
	}

	h.respData(c, data)
}

func (h *Handler) UpdateBot(c *gin.Context) {
	botIDStr := c.Param("bot_id")
	botId, err := strconv.ParseUint(botIDStr, 10, 64)
	if err != nil {
		h.respErr(c, http.StatusBadRequest, err)
		return
	}
	var req api.UpdateBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respErr(c, http.StatusBadRequest, err)
		return
	}

	if _, err := h.llms.GetChatModel(req.ChatModel); err != nil {
		h.respErr(c, http.StatusBadRequest, errors.New("chat model does not exist"))
		return
	}
	if req.Middlewares != nil {
		if err := h.middlewareHandler.ValidateConfig(req.Middlewares); err != nil {
			h.respErr(c, http.StatusBadRequest, err)
			return
		}
	}

	m := map[string]any{
		"name":               req.Name,
		"chat_model":         req.ChatModel,
		"prompt":             req.Prompt,
		"boundary_prompt":    req.BoundaryPrompt,
		"context_turn_count": req.ContextTurnCount,
		"temperature":        req.Temperature,
	}
	if req.Middlewares != nil {
		v := models.MiddlewareConfig(*req.Middlewares)
		m["middlewares"] = &v
	}

	rowsAffected, err := h.sh.UpdateBot(c, uint(botId), m)
	if err != nil {
		h.respErr(c, http.StatusInternalServerError, err)
		return
	}

	if rowsAffected == 0 {
		h.respErr(c, http.StatusNotFound, errors.New("bot not found"))
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteBot(c *gin.Context) {
	botIDStr := c.Param("bot_id")
	botId, err := strconv.ParseUint(botIDStr, 10, 64)
	if err != nil {
		h.respErr(c, http.StatusBadRequest, err)
		return
	}

	if err := h.sh.DeleteBot(c, uint(botId)); err != nil {
		h.respErr(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}
