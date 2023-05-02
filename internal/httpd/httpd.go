package httpd

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pandodao/botastic/config"
)

type Server struct {
	cfg    config.HttpdConfig
	engine *gin.Engine
	h      *Handler
}

func New(cfg config.HttpdConfig, h *Handler) *Server {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.Default()
	s := &Server{
		cfg:    cfg,
		engine: engine,
		h:      h,
	}

	s.initRoutes()
	return s
}

func (s *Server) initRoutes() {
	h := s.h

	s.engine.LoadHTMLGlob("templates/*.html")
	s.engine.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	s.engine.GET("/hc", s.h.HealthCheck)
	v1 := s.engine.Group("/api/v1")
	{
		v1.GET("/models", h.ListModels)

		convs := v1.Group("/conversations")
		{
			convs.POST("/", h.CreateConv)
			convs.GET("/:conv_id", h.GetConv)
			convs.PUT("/:conv_id", h.GetConv)
			convs.DELETE("/:conv_id", h.DeleteConv)
			convs.POST("/:conv_id", h.CreateTurn)
			// conversations.POST("/oneway", conv.CreateOnewayConversation(s.convz, s.convs, s.hub))
			// conversations.GET("/{conversationID}/{turnID}", conv.GetConversationTurn(s.botz, s.convs, s.hub))
		}

		bots := v1.Group("/bots")
		{
			bots.POST("/", h.CreateBot)
			bots.GET("/:bot_id", h.GetBot)
			bots.GET("/", h.GetBots)
			bots.PUT("/:bot_id", h.UpdateBot)
			bots.DELETE("/:bot_id", h.DeleteBot)
		}
	}
}

func (s *Server) Start() error {
	return s.engine.Run(s.cfg.Addr)
}