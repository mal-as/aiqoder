package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
	server *http.Server
	logger *slog.Logger
	cfg    *config
}

func NewServer(opts ...OptionFunc) *Server {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.debugGin {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	srv := &Server{
		router: gin.New(),
		logger: cfg.logger,
		cfg:    cfg,
	}
	srv.router.RemoveExtraSlash = cfg.removeExtraSlash
	srv.router.Use(gin.Recovery())
	srv.router.NoRoute(srv.handlerNotFoundPage)

	srv.server = &http.Server{
		Addr:              cfg.listen,
		Handler:           srv.router,
		ReadTimeout:       cfg.readTimeout,
		ReadHeaderTimeout: cfg.readHeaderTimeout,
		WriteTimeout:      cfg.writeTimeout,
		IdleTimeout:       cfg.idleTimeout,
		MaxHeaderBytes:    cfg.maxHeaderBytes,
	}

	return srv
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

func (s *Server) Server() *http.Server {
	return s.server
}

func (s *Server) handlerNotFoundPage(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"message": "Not Found"})
}

func (s *Server) Start() {
	s.logger.Info("http server listening on http://" + s.cfg.listen)
	defer s.logger.Info("http server closed")

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("http server closed with error:", slog.Attr{
			Key:   "error",
			Value: slog.StringValue(err.Error()),
		})
	}
}

func (s *Server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.gracefulShutdown)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("http server shutdown error:", slog.Attr{
			Key:   "error",
			Value: slog.StringValue(err.Error()),
		})
	}

	s.logger.Info("http server shutdown gracefully")
}
