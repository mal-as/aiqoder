package logging

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		entry := log.With(
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.String()),
		)

		logRW := newLoggingRW(c.Writer)
		c.Writer = logRW
		t1 := time.Now()
		c.Next()

		entry.Info("request completed", []any{
			slog.Int("status", logRW.statusCode),
			slog.String("duration", time.Since(t1).String()),
		}...)
	}
}

type loggingRW struct {
	gin.ResponseWriter
	statusCode int
}

func (rw *loggingRW) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *loggingRW) Write(data []byte) (int, error) {
	return rw.ResponseWriter.Write(data) //nolint:wrapcheck // pass-through delegation to the embedded ResponseWriter
}

func newLoggingRW(writer gin.ResponseWriter) *loggingRW {
	return &loggingRW{
		ResponseWriter: writer,
	}
}
