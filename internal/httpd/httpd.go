package httpd

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pandodao/botastic/config"
	"go.uber.org/zap"
)

type Server struct {
	cfg    config.HttpdConfig
	engine *gin.Engine
	h      *Handler
	logger *zap.Logger
}

func New(cfg config.HttpdConfig, h *Handler, logger *zap.Logger) *Server {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	s := &Server{
		cfg:    cfg,
		engine: gin.New(),
		h:      h,
		logger: logger.Named("httpd"),
	}

	s.initRoutes()
	return s
}

func (s *Server) initRoutes() {
	h := s.h

	s.engine.Use(loggerMiddleware(s.logger), gin.Recovery())
	s.engine.LoadHTMLGlob("templates/*.html")
	s.engine.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	s.engine.GET("/hc", s.h.HealthCheck)
	v1 := s.engine.Group("/api/v1")
	{
		v1.GET("/models", h.ListModels)
		v1.GET("/middlewares", h.ListMiddlewares)

		convs := v1.Group("/conversations")
		{
			convs.POST("/", h.CreateConv)
			convs.GET("/:conv_id", h.GetConv)
			convs.PUT("/:conv_id", h.UpdateConv)
			convs.DELETE("/:conv_id", h.DeleteConv)
			convs.POST("/:conv_id", h.CreateTurn)
			convs.POST("/oneway", h.CreateTurnOneway) // legacy, use POST /turns instead
		}

		turns := v1.Group("/turns")
		{
			turns.POST("/", h.CreateTurnOneway)
			turns.GET("/:turn_id", h.GetTurn)
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

func (s *Server) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    s.cfg.Addr,
		Handler: s.engine,
	}

	go func() {
		s.logger.Info("httpd server started", zap.String("addr", s.cfg.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("failed to listen and serve", zap.Error(err))
		}
	}()
	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}

func loggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		end := time.Now()
		latency := end.Sub(start)

		logger.Info("HTTP request",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", latency),
		)
	}
}
