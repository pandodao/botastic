package httpd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pandodao/botastic/api"
	"github.com/pandodao/botastic/internal/vector"
	llmsapi "github.com/pandodao/botastic/pkg/llms/api"
)

func (h *Handler) UpsertIndexes(c *gin.Context) {
	var req api.UpsertIndexesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respErr(c, http.StatusBadRequest, err)
		return
	}

	resp, err := h.vih.UpsertIndexes(c, req)
	if err != nil {
		if errors.Is(err, llmsapi.ErrModelNotFound) ||
			errors.Is(err, llmsapi.ErrTooManyRequestTokens) {
			h.respErr(c, http.StatusBadRequest, err)
			return
		}

		var verr *vector.IndexNotFoundError
		if errors.As(err, &verr) {
			h.respErr(c, http.StatusBadRequest, err)
			return
		}

		h.respErr(c, http.StatusInternalServerError, fmt.Errorf("failed to upsert indexes: %w", err))
		return
	}

	h.respData(c, resp)
}

func (h *Handler) SearchIndexes(c *gin.Context) {
	var req api.SearchIndexesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respErr(c, http.StatusBadRequest, err)
		return
	}

	result, err := h.vih.SearchIndexes(c, req.EmbeddingModel, req.Keyword, req.GroupKey, req.Limit)
	if err != nil {
		if errors.Is(err, llmsapi.ErrModelNotFound) ||
			errors.Is(err, llmsapi.ErrTooManyRequestTokens) {
			h.respErr(c, http.StatusBadRequest, err)
			return
		}
		h.respErr(c, http.StatusInternalServerError, fmt.Errorf("failed to search indexes: %w", err))
		return
	}

	h.respData(c, result)
}
