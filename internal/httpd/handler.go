package httpd

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pandodao/botastic/api"
	"github.com/pandodao/botastic/internal/llms"
	"github.com/pandodao/botastic/storage"
)

type Handler struct {
	llms *llms.Hanlder
	sh   *storage.Handler
}

func NewHandler(sh *storage.Handler, llms *llms.Hanlder) *Handler {
	return &Handler{
		llms: llms,
		sh:   sh,
	}
}

func (h *Handler) respErr(c *gin.Context, statusCode int, err error, codes ...api.ErrorCode) {
	code := statusCode
	if len(codes) > 0 {
		code = int(codes[0])
	}
	c.JSON(code, api.NewErrorResponse(code, err.Error()))
}

func (h *Handler) respData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, api.NewSuccessResponse(data))
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.String(200, "OK")
}

func (h *Handler) ListModels(c *gin.Context) {
	h.respData(c, api.ListModelsResponse{
		ChatModels:      h.llms.ChatModels(),
		EmbeddingModels: h.llms.EmbeddingModels(),
	})
}
